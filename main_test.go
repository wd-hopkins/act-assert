package act_assert_test

import (
	"github.com/nektos/act/pkg/model"
	//log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wd-hopkins/act-assert"
	"testing"
)

var workflow *act_assert.ActAssert

func init() {
	//log.SetLevel(log.DebugLevel)
	workflow = act_assert.New().WithWorkflowPath("./.github/workflows/example.yaml")
}

func Test_fail_main(t *testing.T) {
	workflow, err := workflow.
		WithEvent("workflow_call").
		Plan()
	assert.NoError(t, err)

	workflow.SetJobResultsFunc(act_assert.Skipped, func(job *model.Job) bool {
		return job.Name != "setup" && job.Name != "job2"
	})

	workflow.Job("setup").SetStepResultsFunc(act_assert.Skipped, func(step *model.Step) bool {
		return step.ID != "setup"
	})

	workflow.Job("job2").
		SetStepResultsFunc(act_assert.Skipped, func(step *model.Step) bool {
			return step.Type() == model.StepTypeUsesActionRemote && step.Name != "Checkout"
		})

	err = workflow.Execute()
	assert.Nil(t, err)

	results := act_assert.NewResults(*workflow)
	assert.Equal(t, act_assert.Success, results.Job("setup").Result())
	assert.Equal(t, act_assert.Success, results.Job("job2").Result())
	logs := results.Job("setup").Logs()
	assert.NotEmpty(t, logs)
	logs = results.Job("job2").Logs()
	assert.NotEmpty(t, logs)
}

func Test_skip_main(t *testing.T) {
	workflow, err := workflow.Plan()
	assert.NoError(t, err)

	workflow.Job("main").
		SetResult(act_assert.Skipped)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	assert.True(t, results.Job("main").Skipped())
	assert.True(t, results.Job("cleanup").Failed())
}

func Test_outputs(t *testing.T) {
	workflow, err := workflow.Plan()
	assert.NoError(t, err)

	workflow.Job("main").
		SetOutput("greeting", "Goodbye!")

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	assert.True(t, results.Job("main").Succeeded())
	assert.True(t, results.Job("cleanup").Failed())
}
