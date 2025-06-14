package act_assert_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	act_assert "github.com/wd-hopkins/act-assert"
)

func Test_override_job_outputs(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath(".github/workflows/example.yaml").
		Plan()
	assert.NoError(t, err)

	workflow.Job("main").
		SetOutput("greeting", "Goodbye!")

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	logs := results.Job("cleanup").Step("Clean up").Logs()
	assert.Equal(t, `The output from the main job was Goodbye!`, logs)
}

func Test_override_step_outputs(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath(".github/workflows/example.yaml").
		Plan()
	assert.NoError(t, err)

	workflow.Job("main").
		Step("output").
		SetOutput("greeting", "Goodbye!")

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	logs := results.Job("cleanup").Step("Clean up").Logs()
	assert.Equal(t, `The output from the main job was Goodbye!`, logs)
}
