package act_assert

import "github.com/nektos/act/pkg/model"

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
