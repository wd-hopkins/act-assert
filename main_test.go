package act_assert_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wdhopkins/act_assert"
	"testing"
)

var workflow *act_assert.ActAssert

func init() {
	//log.SetLevel(log.DebugLevel)
	workflow = act_assert.New().WithWorkflowPath("./.github/workflows/example.yaml")
}

func Test_fail_main(t *testing.T) {
	workflow, err := workflow.Plan()
	assert.NoError(t, err)

	workflow.Job("main").
		SetResult(act_assert.Failure)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	assert.True(t, results.Job("main").Failed())
	assert.True(t, results.Job("cleanup").Skipped())
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
