package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestValidOrg(t *testing.T) {
	t.Parallel()

	// define a project name and validate it in the response
	orgName := uuid.NewString()
	moduleName := "./modules/single_org"

	// copy provider.tf
	defer os.Remove(fmt.Sprintf("%s/provider.tf", moduleName))
	files.CopyFile("./modules/provider.tf", fmt.Sprintf("%s/provider.tf", moduleName))

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: moduleName,
		Vars: map[string]interface{}{
			"org_name": orgName,
		},
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
}
