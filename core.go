package act_assert

import (
	"context"
	"fmt"
	"github.com/nektos/act/pkg/container"
	"github.com/nektos/act/pkg/model"
	"github.com/nektos/act/pkg/runner"
	"os"
	"strings"
)

type ActAssert struct {
	event            string
	jobName          string
	workflowFilePath string
	plan             *model.Plan
	runContexts      []*runner.RunContext
	inputs           map[string]string
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

func (a *ActAssert) WithInputs(inputs map[string]string) *ActAssert {
	if a.inputs == nil {
		a.inputs = make(map[string]string)
	}
	for k, v := range inputs {
		a.inputs[k] = v
	}
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
	r, err := runner.New(&runner.Config{
		GitHubInstance: "github.com",
		RemoteName:     "origin",
		Workdir:        ".",
		DefaultBranch:  "main",
		LogOutput:      true,
		Token:          os.Getenv("GITHUB_TOKEN"),
		Platforms: map[string]string{
			"ubuntu-latest": "node:16-buster-slim",
			"ubuntu-24.04":  "node:16-buster-slim",
		},
		ContainerArchitecture: "linux/amd64",
		ContainerDaemonSocket: socket.Socket,
		NoSkipCheckout:        false,
		ContainerNetworkMode:  "host",
		Inputs:                a.inputs,
		EventName:             a.event,
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
	name        string
	jobRun      *model.Run
	stepOutputs map[string]map[string]string
}

func (j *JobPlan) SetResult(result Result) *JobPlan {
	j.jobRun.Job().Result = string(result)
	return j
}

func (j *JobPlan) SetOutput(k, v string) *JobPlan {
	j.jobRun.Job().Outputs[k] = v
	return j
}

func (j *JobPlan) SetStepResultsFunc(result Result, f func(*model.Step) bool) *JobPlan {
	j.jobRun.StepResultsFunc = func(step *model.Step) (bool, string) {
		return f(step), string(result)
	}
	return j
}

func (j *JobPlan) SetStepOutputs() *JobPlan {
	j.jobRun.StepOutputsFunc = func(step *model.Step) map[string]string {
		for stepName, outputs := range j.stepOutputs {
			if step.ID == stepName || step.Name == stepName {
				return outputs
			}
		}
		return nil
	}
	return j
}

func (j *JobPlan) Step(name string) *StepPlan {
	return &StepPlan{
		name:    name,
		step:    j.jobRun.Job().GetStep(name),
		jobPlan: j,
	}
}

type StepPlan struct {
	name    string
	step    *model.Step
	jobPlan *JobPlan
}

func (s *StepPlan) SetResult(result Result) *StepPlan {
	s.step.Result = string(result)
	return s
}

func (s *StepPlan) SetOutputs(o map[string]string) *StepPlan {
	s.jobPlan.stepOutputs[s.name] = o
	s.jobPlan.SetStepOutputs()
	return s
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

func (j *JobResults) Result() Result {
	return Result(j.runContext.Run.Job().Result)
}

func (j *JobResults) Logs() string {
	if j.runContext.ChildContexts == nil {
		return aggregateStepLogs(j.runContext)
	}
	return aggregateStepLogs((*j.runContext.ChildContexts)[0])
}

//func aggregateJobLogs(runContext *runner.RunContext) string {
//	logs := ""
//	for i, childContext := range *runContext.ChildContexts {
//
//	}
//}

func aggregateStepLogs(runContext *runner.RunContext) string {
	var expressionEvaluator = runContext.NewExpressionEvaluator(context.Background())
	logs := ""
	for _, step := range runContext.Run.Job().Steps {
		if step.Logs != "" {
			logs += prependName(step.Logs, expressionEvaluator.Interpolate(context.Background(), step.Name))
		}
	}
	return logs
}

func prependName(logs, name string) string {
	lines := strings.Split(logs, "\n")
	for i, line := range lines {
		lines[i] = fmt.Sprintf("%s: %s", name, line)
	}
	return strings.Join(lines, "\n")
}
