package act_assert

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wd-hopkins/act/pkg/container"
	"github.com/wd-hopkins/act/pkg/model"
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

func (j *JobPlan) Skip() *JobPlan {
	j.SetResult(Skipped)
	return j
}

func (j *JobPlan) SetOutput(k, v string) *JobPlan {
	if j.jobRun.Job().Outputs == nil {
		j.jobRun.Job().Outputs = make(map[string]string)
	}
	j.jobRun.Job().Outputs[k] = v
	return j
}

func (j *JobPlan) SetStepResultsFunc(result Result, f func(*model.Step) bool) *JobPlan {
	j.jobRun.StepResultsFunc = func(step *model.Step) (bool, string) {
		return f(step), string(result)
	}
	return j
}

func (j *JobPlan) SetContainerImage(image string) *JobPlan {
	j.jobRun.Job().ContainerImageOverride = image
	return j
}

func (j *JobPlan) CopyFileToContainer(hostPath, destPath string) *JobPlan {
	file, err := os.ReadFile(hostPath)
	if err != nil {
		panic(err)
	}
	if j.jobRun.FileMounts == nil {
		j.jobRun.FileMounts = map[string]*container.FileEntry{}
	}
	j.jobRun.FileMounts[filepath.Dir(destPath)] = &container.FileEntry{
		Name: filepath.Base(destPath),
		Mode: 0o644,
		Body: string(file),
	}
	return j
}

func (j *JobPlan) WithBindMount(source, dest string) *JobPlan {
	j.jobRun.BindMounts = append(j.jobRun.BindMounts, fmt.Sprintf("%v:%v", source, dest))
	return j
}

// SetWorkflowName sets the name of the workflow that the job is part of. Applies to all jobs in the workflow.
func (j *JobPlan) SetWorkflowName(name string) *JobPlan {
	j.jobRun.Workflow.Name = name
	return j
}

func (j *JobPlan) setStepOutputs() *JobPlan {
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

// Skip the execution of the step. If setResult is true, the step result will be marked as Skipped.
// Otherwise, the step result can be Success or Skipped depending on the evaluation of the step's 'if' condition.
func (s *StepPlan) Skip(setResult bool) *StepPlan {
	if setResult {
		s.SetResult(Skipped)
	} else {
		s.step.SkipExecution = true
	}
	return s
}

func (s *StepPlan) SetOutputs(o map[string]string) *StepPlan {
	for k, v := range o {
		if s.jobPlan.stepOutputs[s.name] == nil {
			s.jobPlan.stepOutputs[s.name] = map[string]string{}
		}
		s.jobPlan.stepOutputs[s.name][k] = v
	}
	s.jobPlan.setStepOutputs()
	return s
}

func (s *StepPlan) SetOutput(k, v string) *StepPlan {
	return s.SetOutputs(map[string]string{k: v})
}

func (s *StepPlan) SetEnv(envs map[string]string) *StepPlan {
	if s.step.EnvOverrides == nil {
		s.step.EnvOverrides = map[string]string{}
	}
	for k, v := range envs {
		s.step.EnvOverrides[k] = v
	}
	return s
}
