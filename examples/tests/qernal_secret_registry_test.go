package tests

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestRegistrySecret(t *testing.T) {
	t.Parallel()
	randomToken := generateRandomToken()
	envs := checkRequiredEnvs(t, []string{"QERNAL_TOKEN", "QERNAL_PROJECT"})
	qernalToken := envs["QERNAL_TOKEN"]
	projectID := envs["QERNAL_PROJECT"]
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../resources/qernal_secret/registry",
		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"project_id":   projectID,
			"auth_token":   randomToken,
			"qernal_token": qernalToken,
		},
	})
	// Clean up resources with "terraform destroy" at the end of the test
	defer terraform.Destroy(t, terraformOptions)
	// Run "terraform init" and "terraform apply"
	terraform.InitAndApply(t, terraformOptions)
	// Run `terraform output` to get the values of output variables
	registryCreatedAt := terraform.Output(t, terraformOptions, "registry_created_at")
	registryRevision := terraform.Output(t, terraformOptions, "registry_revision")
	// Verify that the creation date is a valid date
	_, err := time.Parse(time.RFC3339, registryCreatedAt)
	assert.NoError(t, err, "registry_created_at should be a valid date")
	// Verify that the revision is a number
	assert.Regexp(t, "^[0-9]+$", registryRevision, "registry_revision should be a number")
	// Verify the name of the created resource
	name := terraform.Output(t, terraformOptions, "qernal_secret_registry_name")
	assert.Equal(t, "my-registry", name)
	// Verify the registry URL of the created resource
	registryURL := terraform.Output(t, terraformOptions, "qernal_secret_registry_registry_url")
	assert.Equal(t, "ghcr.io", registryURL)
}
