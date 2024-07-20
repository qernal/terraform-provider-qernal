package tests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentSecret(t *testing.T) {
	t.Parallel()

	envs := checkRequiredEnvs(t, []string{"QERNAL_TOKEN", "QERNAL_PROJECT"})

	qernalToken := envs["QERNAL_TOKEN"]
	projectID := envs["QERNAL_PROJECT"]
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../resources/qernal_secret/environment",
		Vars: map[string]interface{}{
			"project_id":   projectID,
			"qernal_token": qernalToken,
		},
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Test outputs
	envRevision := terraform.Output(t, terraformOptions, "env_revision")

	// Verify that the revision is a number
	assert.Regexp(t, "^[0-9]+$", envRevision, "env_revision should be a number")

	// Verify the name of the created resource
	name := terraform.Output(t, terraformOptions, "qernal_secret_environment_name")
	assert.Equal(t, "PORT", name)

	// Verify the value of the created resource
	value := terraform.Output(t, terraformOptions, "qernal_secret_environment_value")
	assert.Equal(t, "8080", value)

	// Verify the reference format
	// TODO: uncomment when new release is made

	// reference := terraform.Output(t, terraformOptions, "qernal_secret_environment_reference")
	// expectedReference := fmt.Sprintf("projects:%s/PORT", projectID)
	// assert.Equal(t, expectedReference, reference, "Reference should be in the format projects:PROJECT_ID/SECRET_NAME")
}
