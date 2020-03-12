package tester

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Run enumerates over each subfolder in the provided directory
// stubs out a provider.tf with a fully populated aws provider with the
// provided services or by default it will add all known services then
// executes the first file ending in _test.go against moto_server on the
// first available port
func Run(dir string, env map[string]string, services []string) error {
	vars := mapToKeyValueSlice(mapMerge(
		map[string]string{
			"AWS_ACCESS_KEY_ID":     "mock_access_key",
			"AWS_SECRET_ACCESS_KEY": "mock_secret_key",
			"AWS_REGION":            "us-east-1",
		},
		env,
	))

	jobs, err := buildJobs(dir, vars)
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			j.Err = j.run(services)
			wg.Done()
		}()
	}

	wg.Wait()

	var failed bool
	for _, j := range jobs {
		j.cleanup()
		if j.Err != nil {
			failed = true
			fmt.Printf("[%s]: failed with error: %v\n", j.Name, j.Err)
			continue
		}
		fmt.Printf("[%s]: succeeded\n", j.Name)
	}

	if failed {
		os.Exit(1)
	}
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
	TestBinary   string
}

const urlFmt = "http://localhost:%d"

//nolint: gosec, gocyclo
func (j *job) run(services []string) error {
	port, err := getPort()
	if err != nil {
		return fmt.Errorf("[%s]: %v", j.Name, err)
	}
	j.Env = append(j.Env, "MOTO_PORT="+strconv.Itoa(port))

	moto, err := j.startProcess("moto_server", "-p", strconv.Itoa(port))
	if err != nil {
		return fmt.Errorf("[%s]: %v", j.Name, err)
	}
	defer func() {
		err := kill(moto)
		if err != nil {
			fmt.Printf("failed to terminate process: %v", err)
		}
	}()

	err = writeProvider(j.ProviderFile, services, port)
	if err != nil {
		return fmt.Errorf("[%s]: %v", j.Name, err)
	}

	fmt.Printf("[%s]: waiting for moto to start...\n", j.Name)
	maxRetries := 20
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)
		//nolint: gosec
		//we own all variables related to this url variable
		url := fmt.Sprintf(urlFmt, port)
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

	err = j.runTerraform()
	if err != nil {
		return fmt.Errorf("[%s]: %v", j.Name, err)
	}

	jj := &job{Name: "random", Env: []string{"TESTVAR=test"}, Stderr: os.Stderr, Stdout: os.Stdout}
	testPath := filepath.Join(j.RootPath, "random_test.go")

	testCmd, err := jj.startProcess("go", "test", "-v", testPath)
	if err != nil {
		return fmt.Errorf("failed to start random test: %v", err)
	}
	err = testCmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait on random test: %v", err)
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
	fileName := j.Name
	if runtime.GOOS == "windows" {
		fileName = j.Name + ".exe"
	}

	j.TestBinary = filepath.Join(j.Path, fileName)
	cmd, err := j.startProcess("go", "test", "-v", "-c", "-o", j.TestBinary, j.TestFile)
	if err != nil {
		return fmt.Errorf("[%s]: failed to compile test: %v", j.Name, err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("[%s]: failed to wait for compilation of test: %v", j.Name, err)
	}

	err = retrier(100*time.Millisecond, 10, func() error {
		_, err := os.Stat(j.TestBinary)
		return err
	})
	if err != nil {
		return fmt.Errorf("[%s]: failed to wait for test binary to be written to disk: %s", j.Name, j.TestBinary)
	}

	cmd, err = j.startProcess(j.TestBinary, "-test.v")
	if err != nil {
		return fmt.Errorf("[%s]: %v", j.Name, err)
	}
	return cmd.Wait()
}

func (j *job) cleanup() {
	err := retrier(100*time.Millisecond, 5, func() error {
		return os.Remove(j.ProviderFile)
	})
	if err != nil {
		fmt.Printf("failed to cleanup: %s\n", j.ProviderFile)
	}

	jobDir := filepath.Dir(j.ProviderFile)
	tfstate := filepath.Join(jobDir, "terraform.tfstate")
	tfdir := filepath.Join(jobDir, ".terraform")

	err = retrier(100*time.Millisecond, 5, func() error {
		return os.Remove(tfstate)
	})
	if err != nil {
		fmt.Printf("failed to cleanup: %s\n", tfstate)
	}

	err = retrier(100*time.Millisecond, 5, func() error {
		return os.Remove(j.TestBinary)
	})
	if err != nil {
		fmt.Printf("failed to cleanup: %s\n", tfdir)
	}

	err = retrier(100*time.Millisecond, 5, func() error {
		return os.RemoveAll(tfdir)
	})
	if err != nil {
		fmt.Printf("failed to cleanup directory: %s\n", tfdir)
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
		"darwin":  exec.Command("kill", pid),
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
