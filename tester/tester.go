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

// Config is used to configure how Tester is executed
type Config struct {
	// Dir is the directory that houses all of the test directories
	Dir string

	// Env is a map of environment variables that need to be provided
	// to each executable by default AWS_REGION, AWS_ACCESS_KEY_ID, and
	// AWS_SECRET_ACCESS_KEY are provided with dummy values (and us-east-1)
	Env map[string]string

	// Services is a list of Terraform AWS Provider custom endpoints
	// https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html
	// By default all custom endpoints are statically provided but this
	// provides an option for explicitly listing them
	Services []string

	// JobsPerCPU is the number of jobs that will be executed in parallel
	// per CPU. If it is not set, it will default to 1
	JobsPerCPU int

	// interally used to store parsed ENV variables
	vars []string
}

// Run enumerates over each subfolder in the provided directory
// stubs out a provider.tf with a fully populated aws provider with the
// provided services or by default it will add all known services then
// executes the first file ending in _test.go against moto_server on the
// first available port
func Run(cfg *Config) error {
	// validate and update configuration
	err := prepareConfig(cfg)
	if err != nil {
		return err
	}

	// create job objects from sub-directories
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

	// enumerate for cleanup separately so any lingering printing
	// caused by the interrupt or killing the processes is printed
	// prior to the job report
	for _, j := range jobs {
		j.cleanup()
	}

	// We have either completed all jobs or
	// an interrupt signal has been received
	// print their final status output
	fmt.Printf("\n\n\n\n===== Job Results =====\n")
	var failed bool
	for _, j := range jobs {
		if j.Err != nil {
			failed = true
			fmt.Printf("%-20s%-15s%v\n", j.Name, "FAILED", j.Err)
			continue
		}
		fmt.Printf("%-20s%-15s\n", j.Name, "SUCCESS")
	}

	// we need to exit non-zero if any job failed
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
		// store an empty struct (zero memory alloc)
		// into the free capacity for throttle for each
		// job that we execute, this will block when we
		// reach capacity until a job is completed
		throttle <- struct{}{}

		// waitgroup is needed to prevent the last X
		// jobs from being prematurely killed where X
		// is the capacity of the throttle channel
		wg.Add(1)

		go func() {
			// run the 'j' job and store the error result
			j.Err = j.run(cfg.Services)
			// free one element in the channel
			<-throttle
			// decrement waitgroup by one
			wg.Done()
		}()
	}

	// block until all job go routines have completed
	// this is only hit when we have more throttle capacity
	// than we do jobs remaining
	wg.Wait()

	// all jobs have completed unblock the done channel so
	// we can cleanup, report, and exit
	done <- struct{}{}
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

func (j *job) run(services []string) error {
	port, err := getPort()
	if err != nil {
		return err
	}

	cleanup, err := j.startMoto(port)
	if err != nil {
		return err
	}
	defer cleanup()

	err = writeProvider(j.ProviderFile, services, port)
	if err != nil {
		return err
	}

	err = j.runTerraform()
	if err != nil {
		return err
	}

	return j.runTest()
}

const urlFmt = "http://localhost:%d"

func (j *job) startMoto(port int) (func(), error) {
	// MOTO_PORT should be available as an ENV var to any
	// process started for this job
	j.Env = append(j.Env, "MOTO_PORT="+strconv.Itoa(port))

	moto, err := j.startProcess("moto_server", "-p", strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	fmt.Printf("[%s]: waiting for moto to start...\n", j.Name)
	maxRetries := 20
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)
		//we own all variables related to this url variable
		url := fmt.Sprintf(urlFmt, port)

		var resp *http.Response
		//nolint: gosec
		resp, err := http.Get(url)
		if err == nil {
			err = resp.Body.Close()
			if err != nil {
				fmt.Printf("[%s]: failed to close response body: %v", j.Name, err)
			}
			break
		}
		fmt.Printf("[%s]: waiting for moto to start (attempt %d/%d)\n", j.Name, i+1, maxRetries)
	}

	return func() {
		// if the process has already exited there is no point in
		// continuing, so return to the caller
		if moto.ProcessState != nil &&
			moto.ProcessState.Exited() {
			return
		}
		err = kill(moto)
		if err != nil {
			fmt.Printf("[%s]: failed to terminate process: %v", j.Name, err)
		}
	}, nil
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

	// grab the output pipes so we can prepend the job
	// name before each line so it is easier to discern
	// which message belongs to which job
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

	// keep up with what processes we have started
	// so we can clean them up if we get an interrupt
	// or if something goes badly
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

	// In order to read the output pipes we use a bufio.Scanner
	// which by default reads a line on each pass so we start
	// our wrapping func in a go routine for each pipe
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go j.wrapOutput(j.Stdout, stdoutScanner, wg)
	go j.wrapOutput(j.Stderr, stderrScanner, wg)

	// block until the process has stopped writing to the pipe
	wg.Wait()

	// if something goes wrong return the error to the caller
	if err := stdoutScanner.Err(); err != nil {
		return fmt.Errorf("stdoutScanner failed: %v", err)
	}
	if err := stderrScanner.Err(); err != nil {
		return fmt.Errorf("stderrScanner failed: %v", err)
	}

	return nil
}

func (j *job) wrapOutput(out io.Writer, s *bufio.Scanner, wg *sync.WaitGroup) {
	// Scan pulls one line from the pipe
	// so we can wrap it with the job name
	for s.Scan() {
		_, err := fmt.Fprintf(out, "[%s]: %s\n", j.Name, s.Text())
		if err != nil {
			fmt.Printf("failed to Fprintf: %v\n", err)
		}
	}

	// decrement the waitgroup
	wg.Done()
}

func ignoreNotExistsErr(err error) error {
	if os.IsNotExist(err) {
		return nil
	}
	return err
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
	// make a third map and join the two maps
	// allowing m2 to overwrite anything in m1
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
	// resolve the absolute path for the
	// user provided directory
	base, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path to %s -> %v", dir, err)
	}

	var jobs []*job

	// walk the directory tree for the absolute path
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path: %q -> %v", path, err)
		}
		// if it is not a directory or it is the current
		// directory, then skip it
		if !info.IsDir() || base == path {
			return nil
		}

		// grab all files inside the directory
		// and return only files ending in _test.go
		pattern := filepath.Join(path, "*_test.go")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to list files in: %q -> %v", path, err)
		}

		// if no _test.go files were found then
		// this folder does not qualify as a job
		if len(matches) == 0 {
			return nil
		}

		j := &job{
			// use the last element of the path
			// as the job name
			Name:         filepath.Base(path),
			RootPath:     base,
			Path:         path,
			TestFile:     matches[0], // use the first _test.go file
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
	// open a connection on any free ephemeral port
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	// grab the port used to open that connection
	// and return it to the caller
	if addr, ok := l.Addr().(*net.TCPAddr); ok {
		return addr.Port, nil
	}

	return 0, fmt.Errorf("failed to get an available port")
}

// kill is used to terminate a process that has sub-processes
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

// writeProvider writes the provider.tf file for each job
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

// defaultServices contains the full list of working services from
// https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html
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

// getProvider returns the template formatted contents for the provider.tf
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
	// add the port to the template data
	tmpl = fmt.Sprintf(tmpl, port)

	// create a new template and parse the data
	t, err := template.New("provider").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template string: %v", err)
	}

	// execute the template against a buffer so
	// we can catch any errors
	var byt bytes.Buffer
	err = t.Execute(&byt, serviceList)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template string: %v", err)
	}

	return byt.Bytes(), nil
}
