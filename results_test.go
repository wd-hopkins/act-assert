package act_assert_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	act_assert "github.com/wd-hopkins/act-assert"
)

func Test_get_job_result(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath(".github/workflows/example.yaml").
		Plan()
	assert.NoError(t, err)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	assert.Equal(t, act_assert.Success, results.Job("main").Result())
	assert.Equal(t, act_assert.Failure, results.Job("cleanup").Result())
}

func Test_get_step_logs(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath(".github/workflows/example.yaml").
		Plan()
	assert.NoError(t, err)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	logs := results.Job("cleanup").Step("Clean up").Logs()
	assert.Equal(t, `The output from the main job was 'Hello, nektos/act'`, logs)
}

func Test_get_job_logs(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath(".github/workflows/example.yaml").
		Plan()
	assert.NoError(t, err)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	logs := results.Job("main").Logs()
	assert.Equal(t, `Run a one-line script: Hello, world!
I am nektos/act: Add other actions to build,
I am nektos/act: test, and deploy your project.`, logs)
}

func Test_get_reusable_job_logs(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath("test/caller.yaml").
		Plan()
	assert.NoError(t, err)

	err = workflow.Execute()
	assert.NoError(t, err)

	results := act_assert.NewResults(*workflow)
	logs := results.Job("main").Logs()
	assert.Equal(t, `job_1: Run a one-line script: Hello, world!
job_1: I am job_1: Output from job_1
job_2: Run a one-line script: Hello, world!
job_2: I am job_2: Output from job_2`, logs)
}

func Test_mask_secrets_in_logs(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath(".github/workflows/example.yaml").
		Plan()
	assert.NoError(t, err)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	logs := results.Job("output").Logs()
	assert.NotContains(t, logs, "should-be-masked")
}
