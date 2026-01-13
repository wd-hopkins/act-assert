package act_assert

import (
	"context"
	"github.com/wd-hopkins/act/pkg/artifacts"
	"github.com/wd-hopkins/act/pkg/common"
	"os"
	"strconv"

	"maps"

	"github.com/wd-hopkins/act/pkg/container"
	"github.com/wd-hopkins/act/pkg/model"
	"github.com/wd-hopkins/act/pkg/runner"
)

type ActAssert struct {
	config
	jobName              string
	workflowFilePath     string
	plan                 *model.Plan
	runContexts          []*runner.RunContext
	artifactServerConfig ArtifactServerConfig
}

func New() *ActAssert {
	return &ActAssert{
		config: config{
			gitHubInstance: "github.com",
			remoteName:     "origin",
			workdir:        ".",
			workflowDir:    ".",
			defaultBranch:  "main",
			logOutput:      true,
			token:          os.Getenv("GITHUB_TOKEN"),
			platforms: map[string]string{
				"ubuntu-latest": "node:16-buster-slim",
				"ubuntu-24.04":  "node:16-bullseye-slim",
				"ubuntu-22.04":  "node:16-bullseye-slim",
				"ubuntu-20.04":  "node:16-buster-slim",
				"ubuntu-18.04":  "node:16-buster-slim",
			},
			containerArchitecture: "linux/amd64",
			noSkipCheckout:        false,
			containerNetworkMode:  "host",
			artifactServerAddr:    common.GetOutboundIP().String(),
			artifactServerPort:    "34567",
		},
		workflowFilePath: "./.github/workflows/",
	}
}

func (a *ActAssert) WithWorkflowPath(file string) *ActAssert {
	a.workflowFilePath = file
	return a
}

func (a *ActAssert) WithEvent(event string) *ActAssert {
	a.eventName = event
	return a
}

func (a *ActAssert) WithJobName(name string) *ActAssert {
	a.jobName = name
	return a
}

func (a *ActAssert) WithInputs(inputs map[string]string) *ActAssert {
	if a.inputs == nil {
		a.inputs = make(map[string]string)
	}
	maps.Copy(a.inputs, inputs)
	return a
}

func (a *ActAssert) WithEnvironment(env map[GithubEnv]string) *ActAssert {
	if a.env == nil {
		a.env = make(map[string]string)
	}
	for k, v := range env {
		a.env[string(k)] = v
	}
	return a
}

func (a *ActAssert) WithPlatform(label, image string) *ActAssert {
	a.platforms[label] = image
	return a
}

func (a *ActAssert) WithWorkdir(workdir string) *ActAssert {
	a.workdir = workdir
	return a
}

func (a *ActAssert) WithWorkflowDir(workflowDir string) *ActAssert {
	a.workflowDir = workflowDir
	return a
}

func (a *ActAssert) WithDefaultBranch(branch string) *ActAssert {
	a.defaultBranch = branch
	return a
}

func (a *ActAssert) WithJsonLogger() *ActAssert {
	a.jSONLogger = true
	return a
}

func (a *ActAssert) WithUser(user string) *ActAssert {
	a.runAsUser = user
	return a
}

func (a *ActAssert) WithForcePull(pull bool) *ActAssert {
	a.forcePull = pull
	return a
}

func (a *ActAssert) ConfigureArtifactServer(config ArtifactServerConfig) *ActAssert {
	a.artifactServerPath = config.Path
	if config.Port > 0 {
		a.artifactServerPort = strconv.Itoa(config.Port)
	}
	if config.Host != "" {
		a.artifactServerAddr = config.Host
	}
	return a
}

func (a *ActAssert) Plan() (*ActAssert, error) {
	planner, err := model.NewWorkflowPlanner(a.workflowFilePath, true, false)
	if err != nil {
		return a, err
	}

	if a.jobName != "" {
		a.plan, err = planner.PlanJob(a.jobName)
	} else if a.eventName != "" {
		a.plan, err = planner.PlanEvent(a.eventName)
	} else {
		a.plan, err = planner.PlanAll()
	}

	return a, err
}

func (a *ActAssert) Job(name string) *JobPlan {
	if a.plan == nil {
		panic("Plan is nil. Did you forget to call Plan()?")
	}

	var job *model.Run
	for _, stage := range a.plan.Stages {
		for _, run := range stage.Runs {
			if run.JobID == name {
				job = run
			}
		}
	}

	if job == nil {
		panic("Job not found in plan")
	}
	return &JobPlan{
		jobRun:      job,
		stepOutputs: make(map[string]map[string]string),
	}
}

func (a *ActAssert) AllJobs() []*JobPlan {
	if a.plan == nil {
		panic("Plan is nil. Did you forget to call Plan()?")
	}

	var jobs []*JobPlan

	for _, stage := range a.plan.Stages {
		for _, run := range stage.Runs {
			jobs = append(jobs, &JobPlan{
				name:        run.JobID,
				jobRun:      run,
				stepOutputs: make(map[string]map[string]string),
			})
		}
	}

	return jobs
}

func (a *ActAssert) SetJobResultsFunc(result Result, f func(*model.Job) bool) *ActAssert {
	for _, stage := range a.plan.Stages {
		for _, run := range stage.Runs {
			if f(run.Job()) {
				run.Job().Result = string(result)
			}
		}
	}
	return a
}

func (a *ActAssert) Execute() error {
	socket, err := container.GetSocketAndHost("docker")
	if err != nil {
		return err
	}
	a.containerDaemonSocket = socket.Socket
	r, err := runner.New(a.config.toRunnerConfig())
	if err != nil {
		return err
	}
	ctx := context.Background()

	// Start artifact server if configured
	serverAddr := a.artifactServerAddr
	if serverAddr == "host.docker.internal" {
		serverAddr = "localhost"
	}
	cancel := artifacts.Serve(ctx, a.artifactServerPath, serverAddr, a.artifactServerPort)
	defer func(cancel context.CancelFunc, path string) {
		cancel()
		if a.artifactServerConfig.Cleanup {
			_ = os.RemoveAll(path)
		}
	}(cancel, a.artifactServerPath)

	e := r.NewPlanExecutor(a.plan)
	_ = e(ctx)
	a.runContexts = r.GetRunContexts()
	return nil
}

func (a *ActAssert) Copy() *ActAssert {
	// Create a new ActAssert with the same configuration, save for runContexts
	return &ActAssert{
		config:           a.config.Clone(),
		jobName:          a.jobName,
		workflowFilePath: a.workflowFilePath,
		plan:             a.plan,
	}
}
