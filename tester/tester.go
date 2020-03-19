package tester

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Dir        string
	Env        map[string]string
	vars       []string
	Services   []string
	JobsPerCPU int
}

// Run enumerates over each subfolder in the provided directory
// stubs out a provider.tf with a fully populated aws provider with the
// provided services or by default it will add all known services then
// executes the first file ending in _test.go against moto_server on the
// first available port
func Run(cfg *Config) error {
	err := prepareConfig(cfg)
	if err != nil {
		return err
	}

	jobs, err := buildJobs(cfg.Dir, cfg.vars)
	if err != nil {
		return err
	}

	sigch := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	// listen for Interrupt or Termination signals
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)

	// kick off jobs in a new go routine
	go runJobs(done, cfg, jobs)

	// watch for interrupt signals simultaneously
	go func() {
		sig := <-sigch
		fmt.Printf("interrupt received: %v\n", sig)
		done <- struct{}{} // trigger shutdown of app disrupting jobs
	}()

	<-done // wait for all jobs to complete

	// We have either completed all jobs or
	// an interrupt signal has been received
	// cleanup all jobs and echo their final
	// status output to the screen
	fmt.Printf("\n\n\n\n===== Job Results =====\n")
	var failed bool
	for _, j := range jobs {
		j.cleanup()
		if j.Err != nil {
			failed = true
			fmt.Printf("%-20s%-15s%v\n", j.Name, "FAILED", j.Err)
			continue
		}
		fmt.Printf("%-20s%-15s\n", j.Name, "SUCCESS")
	}

	if failed {
		os.Exit(1)
	}
	return nil
}

func runJobs(done chan struct{}, cfg *Config, jobs []*job) {
	maxProcs := runtime.NumCPU() * cfg.JobsPerCPU
	throttle := make(chan struct{}, maxProcs)
	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		j := j
		throttle <- struct{}{}
		wg.Add(1)
		go func() {
			j.Err = j.run(cfg.Services)
			<-throttle
			wg.Done()
		}()
	}

	wg.Wait()
	done <- struct{}{} // all jobs have completed
}

func prepareConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("a tester.Config must be provided")
	}

	if len(cfg.Dir) == 0 {
		cfg.Dir = "."
	}

	if cfg.JobsPerCPU <= 0 {
		cfg.JobsPerCPU = 1
	}

	cfg.vars = mapToKeyValueSlice(mapMerge(
		map[string]string{
			"AWS_ACCESS_KEY_ID":     "mock_access_key",
			"AWS_SECRET_ACCESS_KEY": "mock_secret_key",
			"AWS_REGION":            "us-east-1",
		},
		cfg.Env,
	))

	return nil
}

type job struct {
	Name         string
	RootPath     string
	Path         string
	TestFile     string
	ProviderFile string
	Env          []string
	Err          error
	Stderr       io.Writer
	Stdout       io.Writer
	Processes    []*exec.Cmd
}

const urlFmt = "http://localhost:%d"

func (j *job) run(services []string) error {
	port, err := getPort()
	if err != nil {
		return err
	}
	j.Env = append(j.Env, "MOTO_PORT="+strconv.Itoa(port))

	moto, err := j.startProcess("moto_server", "-p", strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer func() {
		if moto.ProcessState != nil &&
			moto.ProcessState.Exited() {
			return
		}
		err = kill(moto)
		if err != nil {
			fmt.Printf("[%s]: failed to terminate process: %v", j.Name, err)
		}
	}()

	err = writeProvider(j.ProviderFile, services, port)
	if err != nil {
		return err
	}

	fmt.Printf("[%s]: waiting for moto to start...\n", j.Name)
	maxRetries := 20
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)
		//we own all variables related to this url variable
		url := fmt.Sprintf(urlFmt, port)

		var resp *http.Response
		//nolint: gosec
		resp, err = http.Get(url)
		if err == nil {
			err = resp.Body.Close()
			if err != nil {
				fmt.Printf("[%s]: failed to close response body: %v", j.Name, err)
			}
			break
		}
		fmt.Printf("[%s]: waiting for moto to start (attempt %d/%d)\n", j.Name, i+1, maxRetries)
	}

	err = j.runTerraform()
	if err != nil {
		return err
	}

	return j.runTest()
}

func (j *job) runTerraform() error {
	init, err := j.startProcess("terraform", "init", "-no-color")
	if err != nil {
		return fmt.Errorf("failed to initialize terraform: %v", err)
	}
	err = init.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for terraform initialization: %v", err)
	}

	apply, err := j.startProcess("terraform", "apply", "-auto-approve", "-no-color")
	if err != nil {
		return fmt.Errorf("failed to apply terraform: %v", err)
	}
	err = apply.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for terraform apply: %v", err)
	}
	return err
}

func (j *job) runTest() error {
	cmd, err := j.startProcess("go", "test", "-v", j.TestFile)
	if err != nil {
		return fmt.Errorf("failed to execute test: %v", err)
	}
	return cmd.Wait()
}

func (j *job) cleanup() {
	j.cleanupProcesses()

	err := retrier(100*time.Millisecond, 10, func() error {
		return ignoreNotExistsErr(os.Remove(j.ProviderFile))
	})
	if err != nil {
		fmt.Printf("[%s]: failed to cleanup: %s -> %v\n", j.Name, j.ProviderFile, err)
	}

	jobDir := filepath.Dir(j.ProviderFile)
	tfstate := filepath.Join(jobDir, "terraform.tfstate")
	tflock := filepath.Join(jobDir, ".terraform.tfstate.lock.info")
	tfdir := filepath.Join(jobDir, ".terraform")

	err = retrier(100*time.Millisecond, 10, func() error {
		return ignoreNotExistsErr(os.Remove(tfstate))
	})
	if err != nil {
		fmt.Printf("[%s]: failed to cleanup: %s -> %v\n", j.Name, tfstate, err)
	}

	err = retrier(100*time.Millisecond, 10, func() error {
		return ignoreNotExistsErr(os.Remove(tflock))
	})
	if err != nil {
		fmt.Printf("[%s]: failed to cleanup: %s -> %v\n", j.Name, tflock, err)
	}

	err = retrier(100*time.Millisecond, 10, func() error {
		return ignoreNotExistsErr(os.RemoveAll(tfdir))
	})
	if err != nil {
		fmt.Printf("[%s]: failed to cleanup directory: %s -> %v\n", j.Name, tfdir, err)
	}
}

func ignoreNotExistsErr(err error) error {
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (j *job) cleanupProcesses() {
	for _, p := range j.Processes {
		if p.ProcessState != nil &&
			p.ProcessState.Exited() {
			continue
		}
		err := kill(p)
		if err != nil {
			fmt.Printf("[%s]: failed to kill process: %s -> %v\n", j.Name, p.Path, err)
		}
	}
}

//nolint: errcheck
func (j *job) startProcess(path string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(path, args...)
	cmd.Dir = j.Path
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, j.Env...)

	stdoutP, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout pipe: %v", err)
	}
	stderrP, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start process: %s -> %v", path, err)
	}

	j.Processes = append(j.Processes, cmd)

	go func() {
		j.readOutput(stdoutP, stderrP)
		// this will always fail as the process
		// can only be fully terminated using the
		// kill method, added nolint: errcheck to
		// avoid linting errors related to this
		cmd.Wait()
	}()

	return cmd, nil
}

func (j *job) readOutput(stdout, stderr io.Reader) error {
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go j.wrapOutput(j.Stdout, stdoutScanner, wg)
	go j.wrapOutput(j.Stderr, stderrScanner, wg)
	wg.Wait()

	if err := stdoutScanner.Err(); err != nil {
		return fmt.Errorf("stdoutScanner failed: %v", err)
	}
	if err := stderrScanner.Err(); err != nil {
		return fmt.Errorf("stderrScanner failed: %v", err)
	}

	return nil
}

func (j *job) wrapOutput(out io.Writer, s *bufio.Scanner, wg *sync.WaitGroup) {
	for s.Scan() {
		_, err := fmt.Fprintf(out, "[%s]: %s\n", j.Name, s.Text())
		if err != nil {
			fmt.Printf("failed to Fprintf: %v\n", err)
		}
	}
	wg.Done()
}

//nolint:unparam
func retrier(delay time.Duration, attempts int, fn func() error) (err error) {
	for attempt := 0; attempt < attempts; attempt++ {
		err = fn()
		if err == nil {
			return
		}
		time.Sleep(delay)
	}
	return
}

func mapToKeyValueSlice(m map[string]string) []string {
	if len(m) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, k+"="+v)
	}
	return result
}

func mapMerge(m1, m2 map[string]string) map[string]string {
	m3 := make(map[string]string, len(m1)+len(m2))
	for k := range m1 {
		m3[k] = m1[k]
	}
	for k := range m2 {
		m3[k] = m2[k]
	}
	return m3
}

func buildJobs(dir string, env []string) ([]*job, error) {
	base, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path to %s -> %v", dir, err)
	}

	var jobs []*job
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path: %q -> %v", path, err)
		}
		if !info.IsDir() || base == path {
			return nil
		}

		pattern := filepath.Join(path, "*_test.go")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to list files in: %q -> %v", path, err)
		}

		if len(matches) == 0 {
			return nil
		}

		j := &job{
			Name:         filepath.Base(path),
			RootPath:     base,
			Path:         path,
			TestFile:     matches[0],
			ProviderFile: filepath.Join(path, "provider.tf"),
			Env:          env,
			Err:          errors.New("job not executed"),
			Stderr:       os.Stderr,
			Stdout:       os.Stdout,
		}

		jobs = append(jobs, j)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func getPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	if addr, ok := l.Addr().(*net.TCPAddr); ok {
		return addr.Port, nil
	}

	return 0, fmt.Errorf("failed to get an available port")
}

func kill(cmd *exec.Cmd) error {
	pid := strconv.Itoa(cmd.Process.Pid)
	killers := map[string]*exec.Cmd{
		"darwin":  exec.Command("kill", "-9", pid),
		"linux":   exec.Command("kill", "-9", pid),
		"windows": exec.Command("TASKKILL", "/T", "/F", "/PID", pid),
	}
	if v, ok := killers[runtime.GOOS]; ok {
		return v.Run()
	}
	return fmt.Errorf("runtime %s is not supported yet", runtime.GOOS)
}

func writeProvider(path string, services []string, port int) error {
	data, err := getProvider(services, port)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write provider at: %q -> %v", path, err)
	}
	return nil
}

var defaultServices = []string{
	"accessanalyzer",
	"acm",
	"acmpca",
	"amplify",
	"apigateway",
	"applicationautoscaling",
	"applicationinsights",
	"appmesh",
	"appstream",
	"appsync",
	"athena",
	"autoscaling",
	"autoscalingplans",
	"backup",
	"batch",
	"budgets",
	"cloud9",
	"cloudformation",
	"cloudfront",
	"cloudhsm",
	"cloudsearch",
	"cloudtrail",
	"cloudwatch",
	"cloudwatchevents",
	"cloudwatchlogs",
	"codebuild",
	"codecommit",
	"codedeploy",
	"codepipeline",
	//"codestarnotifications", doesn't work
	"cognitoidentity",
	"cognitoidp",
	"configservice",
	"cur",
	"dataexchange",
	"datapipeline",
	"datasync",
	"dax",
	"devicefarm",
	"directconnect",
	"dlm",
	"dms",
	"docdb",
	"ds",
	"dynamodb",
	"ec2",
	"ecr",
	"ecs",
	"efs",
	"eks",
	"elasticache",
	"elasticbeanstalk",
	"elastictranscoder",
	"elb",
	"emr",
	"es",
	"firehose",
	"fms",
	"forecast",
	"fsx",
	"gamelift",
	"glacier",
	"globalaccelerator",
	"glue",
	"guardduty",
	"greengrass",
	"iam",
	"imagebuilder",
	"inspector",
	"iot",
	"iotanalytics",
	"iotevents",
	"kafka",
	"kinesis",
	"kinesisanalytics",
	"kinesisvideo",
	"kms",
	"lakeformation",
	"lambda",
	"lexmodels",
	"licensemanager",
	"lightsail",
	"macie",
	"managedblockchain",
	"marketplacecatalog",
	"mediaconnect",
	"mediaconvert",
	"medialive",
	"mediapackage",
	"mediastore",
	"mediastoredata",
	"mq",
	"neptune",
	"opsworks",
	"organizations",
	"personalize",
	"pinpoint",
	"pricing",
	"qldb",
	"quicksight",
	"ram",
	"rds",
	"redshift",
	"resourcegroups",
	"route53",
	"route53resolver",
	"s3",
	"s3control",
	"sagemaker",
	"sdb",
	"secretsmanager",
	"securityhub",
	"serverlessrepo",
	"servicecatalog",
	"servicediscovery",
	"servicequotas",
	"ses",
	"shield",
	"sns",
	"sqs",
	"ssm",
	"stepfunctions",
	"storagegateway",
	"sts",
}

func getProvider(services []string, port int) ([]byte, error) {
	serviceList := defaultServices
	if len(services) > 0 {
		serviceList = services
	}

	tmpl := `terraform {
	backend "local" {
		path = "terraform.tfstate"
	}
}

provider "aws" {
	s3_force_path_style         = true
	skip_credentials_validation = true
	skip_metadata_api_check     = true
	skip_requesting_account_id  = true
	endpoints {
{{- range .}}
		{{.}} = "http://localhost:%d"
{{- end}}
	}
}
`
	tmpl = fmt.Sprintf(tmpl, port)
	t, err := template.New("provider").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template string: %v", err)
	}

	var byt bytes.Buffer
	err = t.Execute(&byt, serviceList)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template string: %v", err)
	}

	return byt.Bytes(), nil
}
