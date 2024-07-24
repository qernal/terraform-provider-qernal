package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/terraform"
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
	files.CopyFile("./modules/provider.tf", fmt.Sprintf("%s/provider.tf", moduleName))

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

	// TODO: validate output
}
