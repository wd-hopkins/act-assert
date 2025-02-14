package act_assert

import (
	"context"
	"github.com/nektos/act/pkg/container"
	"github.com/nektos/act/pkg/model"
	"github.com/nektos/act/pkg/runner"
	"os"
)

type ActAssert struct {
	event            string
	jobName          string
	workflowFilePath string
	plan             *model.Plan
	runContexts      []*runner.RunContext
}

func New() *ActAssert {
	return &ActAssert{}
}

func (a *ActAssert) WithWorkflowPath(file string) *ActAssert {
	a.workflowFilePath = file
	return a
}

func (a *ActAssert) WithEvent(event string) *ActAssert {
	a.event = event
	return a
}

func (a *ActAssert) WithJobName(name string) *ActAssert {
	a.jobName = name
	return a
}

func (a *ActAssert) Plan() (*ActAssert, error) {
	planner, err := model.NewWorkflowPlanner(a.workflowFilePath, true)
	if err != nil {
		return a, err
	}

	if a.jobName != "" {
		a.plan, err = planner.PlanJob(a.jobName)
	} else if a.event != "" {
		a.plan, err = planner.PlanEvent(a.event)
	} else {
		a.plan, err = planner.PlanAll()
	}

	return a, err
}

func (a *ActAssert) Job(name string) *JobPlan {
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
		jobRun: job,
	}
}

func (a *ActAssert) Execute() error {
	socket, err := container.GetSocketAndHost("docker")
	if err != nil {
		return err
	}
	r, err := runner.New(&runner.Config{
		GitHubInstance: "github.com",
		RemoteName:     "origin",
		Workdir:        ".",
		DefaultBranch:  "main",
		LogOutput:      true,
		Token:          os.Getenv("GITHUB_TOKEN"),
		Platforms: map[string]string{
			"ubuntu-latest":         "node:16-buster-slim",
			"my-self-hosted-runner": "registry-adapter.tools.cosmic.sky/core-platform/core-engineering/core-platform/core-go-action-runner-1.21:latest",
		},
		ContainerArchitecture: "linux/amd64",
		ContainerDaemonSocket: socket.Socket,
		NoSkipCheckout:        false,
		ContainerNetworkMode:  "host",
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	e := r.NewPlanExecutor(a.plan)
	err = e(ctx)
	a.runContexts = r.GetRunContexts()
	return err
}

type Result string

const (
	Success Result = "success"
	Failure Result = "failure"
	Skipped Result = "skipped"
)

type JobPlan struct {
	name   string
	jobRun *model.Run
}

func (j *JobPlan) SetResult(result Result) *JobPlan {
	j.jobRun.Skip = true
	j.jobRun.Job().Result = string(result)
	return j
}

func (j *JobPlan) SetOutput(k, v string) *JobPlan {
	j.jobRun.Job().Outputs[k] = v
	return j
}

type Results struct {
	runContexts []*runner.RunContext
}

func NewResults(act ActAssert) *Results {
	return &Results{
		runContexts: act.runContexts,
	}
}

func (r *Results) Job(name string) *JobResults {
	for _, ctx := range r.runContexts {
		if ctx.JobName == name {
			return &JobResults{
				JobName:    name,
				runContext: ctx,
			}
		}
	}
	panic("Job not found in results")
	return nil
}

type JobResults struct {
	JobName    string
	runContext *runner.RunContext
}

func (j *JobResults) Succeeded() bool {
	return j.runContext.Run.Job().Result == string(Success)
}

func (j *JobResults) Skipped() bool {
	return j.runContext.Run.Job().Result == string(Skipped)
}

func (j *JobResults) Failed() bool {
	return j.runContext.Run.Job().Result == string(Failure)
}
