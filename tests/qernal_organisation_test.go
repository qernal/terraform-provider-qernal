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

func TestValidOrg(t *testing.T) {
	t.Parallel()

	// define a project name and validate it in the response
	orgName := uuid.NewString()
	moduleName := "./modules/single_org"

	// copy provider.tf
	defer os.Remove(fmt.Sprintf("%s/provider.tf", moduleName))
	err := files.CopyFile("./modules/provider.tf", fmt.Sprintf("%s/provider.tf", moduleName))
	if err != nil {
		t.Error("failed to copy file")
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: moduleName,
		Vars: map[string]interface{}{
			"org_name": orgName,
		},
	})

	defer func() {
		if err := cleanupTerraformFiles(moduleName); err != nil {
			t.Logf("Warning: Failed to clean up Terraform files: %v", err)
		}
	}()

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// validate output
	tfOrgName := terraform.Output(t, terraformOptions, "organisation_name")
	assert.Equal(t, orgName, tfOrgName)
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
