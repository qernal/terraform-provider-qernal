package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentSecret(t *testing.T) {
	t.Parallel()

	orgId, _, err := createOrg()
	if err != nil {
		t.Fatal("Failed to create org")
	}

	projectId, _, err := createProject(orgId)
	if err != nil {
		t.Fatal("Failed to create project")
	}

	secretName := strings.ToUpper(randomSecretName())
	secretValue := uuid.NewString()

	moduleName := "./modules/environment_secret"

	// Copy provider.tf
	defer os.Remove(fmt.Sprintf("%s/provider.tf", moduleName))
	err = files.CopyFile("./modules/provider.tf", fmt.Sprintf("%s/provider.tf", moduleName))
	if err != nil {
		t.Fatal("failed to copy provider file")
	}

	defer func() {
		if err := cleanupTerraformFiles(moduleName); err != nil {
			t.Logf("Warning: Failed to clean up Terraform files: %v", err)
		}
	}()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: moduleName,
		Vars: map[string]interface{}{
			"project_id":   projectId,
			"secret_name":  secretName,
			"secret_value": secretValue,
		},
	})

	defer deleteProject(projectId)
	defer deleteOrg(orgId)
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Validate outputs
	outputSecretName := terraform.Output(t, terraformOptions, "secret_name")
	assert.Equal(t, secretName, outputSecretName)

	outputSecretValue := terraform.Output(t, terraformOptions, "secret_value")
	assert.Equal(t, secretValue, outputSecretValue)
}
