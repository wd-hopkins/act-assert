package act_assert

import (
	"context"
	"os"

	"maps"

	"github.com/nektos/act/pkg/container"
	"github.com/nektos/act/pkg/model"
	"github.com/nektos/act/pkg/runner"
)

type ActAssert struct {
	config
	jobName          string
	workflowFilePath string
	plan             *model.Plan
	runContexts      []*runner.RunContext
}

func New() *ActAssert {
	return &ActAssert{
		config: config{
			gitHubInstance: "github.com",
			remoteName:     "origin",
			workdir:        ".",
			defaultBranch:  "main",
			logOutput:      true,
			token:          os.Getenv("GITHUB_TOKEN"),
			platforms: map[string]string{
				"ubuntu-latest": "node:16-buster-slim",
				"ubuntu-22.04":  "node:16-bullseye-slim",
				"ubuntu-20.04":  "node:16-buster-slim",
				"ubuntu-18.04":  "node:16-buster-slim",
			},
			containerArchitecture: "linux/amd64",
			noSkipCheckout:        false,
			containerNetworkMode:  "host",
		},
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

func (a *ActAssert) Plan() (*ActAssert, error) {
	planner, err := model.NewWorkflowPlanner(a.workflowFilePath, true)
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
	e := r.NewPlanExecutor(a.plan)
	err = e(ctx)
	a.runContexts = r.GetRunContexts()
	return err
}
