package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestValidProject(t *testing.T) {
	t.Parallel()

	// orgId, orgName, err := createOrg()
	orgId, _, err := createOrg()
	if err != nil {
		t.Fatal("Failed to create org")
	}

	// define a project name and validate it in the response
	projectName := uuid.NewString()
	moduleName := "./modules/single_project"

	// copy provider.tf
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
			"org_id":       orgId,
			"project_name": projectName,
		},
	})

	defer deleteOrg(orgId)
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// validate output
	tfProjectName := terraform.Output(t, terraformOptions, "project_name")
	assert.Equal(t, projectName, tfProjectName)
}

func TestOrganisationDataSource(t *testing.T) {
	t.Parallel()

	orgId, _, err := createOrg()
	if err != nil {
		t.Fatal("Failed to create org")
	}

	moduleName := "./modules/org_datasource"

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
			"org_id": orgId,
		},
	})

	defer deleteOrg(orgId)
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Validate output
	tfOrgID := terraform.Output(t, terraformOptions, "organisation_id")
	assert.Equal(t, orgId, tfOrgID)

}
