package act_assert_test

import (
	"github.com/stretchr/testify/assert"
	act_assert "github.com/wd-hopkins/act-assert"
	"testing"
)

func Test_artifact_server(t *testing.T) {
	workflow, err := act_assert.New().
		WithWorkflowPath("test/artifacts.yaml").
		ConfigureArtifactServer(act_assert.ArtifactServerConfig{
			Path:    ".artifacts",
			Host:    "host.docker.internal",
			Cleanup: false,
		}).
		Plan()
	assert.NoError(t, err)
	
	err = workflow.Execute()
	assert.NoError(t, err)

	results := act_assert.NewResults(*workflow)
	assert.Equal(t, act_assert.Success, results.Job("upload").Result())
	downloadJob := results.Job("download")
	assert.Equal(t, act_assert.Success, downloadJob.Result())
	assert.Equal(t, "test-artifact.txt", downloadJob.Step("List files").Logs())
}
