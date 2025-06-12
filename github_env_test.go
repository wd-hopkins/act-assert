package act_assert_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	act_assert "github.com/wd-hopkins/act-assert"
)

func Test_override_github_env(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath("test/vars.yaml").
		WithEnvironment(map[act_assert.GithubEnv]string{
			act_assert.GithubRef: "test-branch",
		}).
		Plan()
	assert.NoError(t, err)

	_ = workflow.Execute()

	results := act_assert.NewResults(*workflow)
	logs := results.Job("env_job").Logs()
	assert.Contains(t, logs, `branch=test-branch`)
}
