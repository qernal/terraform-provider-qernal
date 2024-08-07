package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

// func validateResponseBody(status int, body string) bool {
// 	if status == 200 && strings.Contains(body, "Hello Word") {
// 		return true
// 	}

// 	return false
// }

func TestValidFunction(t *testing.T) {
	t.Parallel()

	// create org
	orgId, _, err := createOrg()
	if err != nil {
		t.Fatal("failed to create org")
	}

	// create project
	projId, _, err := createProj(orgId)
	if err != nil {
		t.Fatal("failed to create project")
	}

	// define a project name and validate it in the response
	functionName := uuid.NewString()
	moduleName := "./modules/single_function"

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
			"project_id":    projId,
			"function_name": functionName,
		},
	})

	defer deleteOrg(orgId)
	defer deleteProj(projId)
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// validate output
	// tfFunctionName := terraform.Output(t, terraformOptions, "function_name")
	// assert.Equal(t, functionName, tfFunctionName)

	// // validate our function deployed
	// host, err := getDefaultHost(projId)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// http_helper.HttpGetWithRetryWithCustomValidation(t, fmt.Sprintf("https://%s/", host), nil, 30, 3*time.Second, validateResponseBody)
}

func TestUpdateFunction(t *testing.T) {
	t.Parallel()

	// create org
	orgId, _, err := createOrg()
	if err != nil {
		t.Fatal("failed to create org")
	}

	// create project
	projId, _, err := createProj(orgId)
	if err != nil {
		t.Fatal("failed to create project")
	}

	// define a project name and validate it in the response
	functionName := uuid.NewString()
	functionNameUpdated := uuid.NewString()
	moduleName := "./modules/single_function_update"

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

	// initial creation
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: moduleName,
		Vars: map[string]interface{}{
			"project_id":    projId,
			"function_name": functionName,
		},
	})

	defer deleteOrg(orgId)
	defer deleteProj(projId)
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// update
	terraformOptionsUpdate := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: moduleName,
		Vars: map[string]interface{}{
			"project_id":    projId,
			"function_name": functionNameUpdated,
		},
	})

	terraform.InitAndApply(t, terraformOptionsUpdate)

	// validate output
	// tfFunctionName := terraform.Output(t, terraformOptions, "function_name")
	// assert.Equal(t, functionName, tfFunctionName)

	// // validate our function deployed
	// host, err := getDefaultHost(projId)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// http_helper.HttpGetWithRetryWithCustomValidation(t, fmt.Sprintf("https://%s/", host), nil, 30, 3*time.Second, validateResponseBody)
}

// func TestValidFunctionWithSecret(t *testing.T) {
// 	t.Parallel()

// 	// create org
// 	orgId, _, err := createOrg()
// 	if err != nil {
// 		t.Fatal("failed to create org")
// 	}

// 	// create project
// 	projId, _, err := createProj(orgId)
// 	if err != nil {
// 		t.Fatal("failed to create project")
// 	}

// 	// define a project name and validate it in the response
// 	functionName := uuid.NewString()
// 	moduleName := "./modules/single_function"

// 	// copy provider.tf
// 	defer os.Remove(fmt.Sprintf("%s/provider.tf", moduleName))
// 	err = files.CopyFile("./modules/provider.tf", fmt.Sprintf("%s/provider.tf", moduleName))
// 	if err != nil {
// 		t.Fatal("failed to copy provider file")
// 	}

// 	defer func() {
// 		if err := cleanupTerraformFiles(moduleName); err != nil {
// 			t.Logf("Warning: Failed to clean up Terraform files: %v", err)
// 		}
// 	}()

// 	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
// 		TerraformDir: moduleName,
// 		Vars: map[string]interface{}{
// 			"project_id":    projId,
// 			"function_name": functionName,
// 		},
// 	})

// 	defer deleteOrg(orgId)
// 	defer deleteProj(projId)
// 	defer terraform.Destroy(t, terraformOptions)

// 	terraform.InitAndApply(t, terraformOptions)

// 	// validate output
// 	// tfFunctionName := terraform.Output(t, terraformOptions, "function_name")
// 	// assert.Equal(t, functionName, tfFunctionName)

// 	// validate our function deployed
// 	host, err := getDefaultHost(projId)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	http_helper.HttpGetWithRetryWithCustomValidation(t, fmt.Sprintf("https://%s/", host), nil, 30, 3*time.Second, validateResponseBody)
// }
