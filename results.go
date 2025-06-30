package act_assert

import (
	"context"
	"fmt"
	"strings"

	"github.com/wd-hopkins/act/pkg/model"
	"github.com/wd-hopkins/act/pkg/runner"
)

type Result string

const (
	Success Result = "success"
	Failure Result = "failure"
	Skipped Result = "skipped"
)

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
		if ctx.JobName == name || ctx.Run.JobID == name {
			return &JobResults{
				JobName:    ctx.JobName,
				runContext: ctx,
			}
		}
		if ctx.ChildContexts != nil {
			for _, childContext := range *ctx.ChildContexts {
				if childContext.JobName == name || childContext.Run.JobID == name {
					return &JobResults{
						JobName:    ctx.JobName,
						runContext: ctx,
					}
				}
			}
		}
	}
	panic(fmt.Sprintf("Job %s not found in results", name))
}

func (r *Results) MatrixJob(name string) MatrixJobResults {
	var matrixResults MatrixJobResults
	for _, ctx := range r.runContexts {
		if ctx.Run.JobID == name {
			matrixResults = append(matrixResults, &JobResults{
				JobName:    ctx.Run.JobID,
				runContext: ctx,
			})
		}
		if ctx.ChildContexts != nil {
			for _, childContext := range *ctx.ChildContexts {
				if childContext.JobName == name || childContext.Run.JobID == name {
					matrixResults = append(matrixResults, &JobResults{
						JobName:    childContext.Run.JobID,
						runContext: childContext,
					})
				}
			}
		}
	}
	if len(matrixResults) <= 0 {
		panic(fmt.Sprintf("Job %s not found in results", name))
	}
	return matrixResults
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

func (j *JobResults) Outputs() map[string]string {
	return j.runContext.Run.Job().Outputs
}

func (j *JobResults) Masks() []string {
	return j.runContext.Masks
}

func (j *JobResults) Logs() string {
	if j.runContext.ChildContexts == nil {
		return aggregateStepLogs(j.runContext)
	}
	return aggregateReusableJobLogs(j.runContext)
}

func (j *JobResults) Step(name string) *StepResults {
	jobType, _ := j.runContext.Run.Job().Type()
	if jobType == model.JobTypeReusableWorkflowLocal ||
		jobType == model.JobTypeReusableWorkflowRemote ||
		j.runContext.ChildContexts != nil {
		panic("Job is calling a reusable workflow and has no steps")
	}
	for _, step := range j.runContext.Run.Job().Steps {
		if step.ID == name || step.Name == name {
			return &StepResults{
				StepName: name,
				step:     step,
			}
		}
	}
	panic("Step not found in Job results")
}

type MatrixJobResults []*JobResults

type StepResults struct {
	StepName string
	step     *model.Step
}

func (s *StepResults) Result() Result {
	return Result(s.step.Result)
}

func (s *StepResults) Logs() string {
	return strings.TrimSpace(s.step.Logs)
}

func aggregateReusableJobLogs(runContext *runner.RunContext) string {
	logs := ""
	for _, childContext := range *runContext.ChildContexts {
		if logs != "" {
			logs += "\n"
		}
		logs += prependName(aggregateStepLogs(childContext), childContext.Name, childContext)
	}
	return logs
}

func aggregateStepLogs(runContext *runner.RunContext) string {
	logs := ""
	for _, step := range runContext.Run.Job().Steps {
		if step.Logs != "" {
			if logs != "" {
				logs += "\n"
			}
			logs += prependName(step.Logs, step.Name, runContext)
		}
	}
	return logs
}

func prependName(logs, name string, runContext *runner.RunContext) string {
	var expressionEvaluator = runContext.NewExpressionEvaluator(context.Background())
	lines := strings.Split(logs, "\n")
	var out []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, fmt.Sprintf("%s: %s", expressionEvaluator.Interpolate(context.Background(), name), line))
		}
	}
	return strings.Join(out, "\n")
}
