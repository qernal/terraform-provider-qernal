package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	openapi_chaos_client "github.com/qernal/openapi-chaos-go-client"
	"golang.org/x/oauth2/clientcredentials"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

var (
	qernalToken         = os.Getenv("QERNAL_TOKEN")
	qernalChaosEndpoint = getEnv("QERNAL_HOST_CHAOS", "https://chaos.qernal.com")
	qernalHydraEndpoint = getEnv("QERNAL_HOST_HYDRA", "https://hydra.qernal.com")
	accessToken, _      = _getAccessToken(qernalToken)
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func _getAccessToken(token string) (string, error) {
	if !strings.Contains(token, "@") || strings.Count(token, "@") > 1 {
		err := errors.New("the qernal token is invalid")
		return "", err
	}

	clientId := strings.Split(token, "@")[0]
	clientSecret := strings.Split(token, "@")[1]

	config := clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", qernalHydraEndpoint),
	}

	oauthToken, err := config.Token(context.TODO())
	if err != nil {
		return "", err
	}

	return oauthToken.AccessToken, nil
}

func qernalClient() *openapi_chaos_client.APIClient {
	configuration := &openapi_chaos_client.Configuration{
		Servers: openapi_chaos_client.ServerConfigurations{
			{
				URL: fmt.Sprintf("%s/v1", qernalChaosEndpoint),
			},
		},
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		},
	}

	apiClient := openapi_chaos_client.NewAPIClient(configuration)
	return apiClient
}

func createOrg() (string, string, error) {
	client := qernalClient()
	organisationBody := *openapi_chaos_client.NewOrganisationBody(uuid.NewString())
	resp, r, err := client.OrganisationsAPI.OrganisationsCreate(context.Background()).OrganisationBody(organisationBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrganisationsAPI.OrganisationsCreate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

		return "", "", err
	}

	return resp.Id, resp.Name, nil
}

func createProject(orgid string) (string, string, error) {

	client := qernalClient()

	projectbody := *openapi_chaos_client.NewProjectBody(orgid, uuid.NewString())

	resp, r, err := client.ProjectsAPI.ProjectsCreate(context.Background()).ProjectBody(projectbody).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.ProjectsCreate`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

		return "", "", err
	}

	return resp.Id, resp.Name, nil
}

func deleteOrg(orgid string) {
	client := qernalClient()
	_, r, err := client.OrganisationsAPI.OrganisationsDelete(context.Background(), orgid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrganisationsAPI.OrganisationsDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}

func createProj(orgid string) (string, string, error) {
	client := qernalClient()
	projectBody := *openapi_chaos_client.NewProjectBody(orgid, uuid.NewString())
	resp, r, err := client.ProjectsAPI.ProjectsCreate(context.Background()).ProjectBody(projectBody).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.ProjectsCreate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

		return "", "", err
	}

	return resp.Id, resp.Name, nil
}

func deleteProj(projid string) {
	client := qernalClient()
	_, r, err := client.ProjectsAPI.ProjectsDelete(context.Background(), projid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.ProjectsDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}

func getDefaultHost(projid string) (string, error) {
	client := qernalClient()
	resp, r, err := client.HostsAPI.ProjectsHostsList(context.Background(), projid).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.ProjectsCreate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

		return "", err
	}

	for _, host := range resp.Data {
		if host.ReadOnly {
			return host.Host, nil
		}
	}

	return "", errors.New("no default host on project")
}

func deleteProject(projectid string) {
	client := qernalClient()

	_, r, err := client.ProjectsAPI.ProjectsDelete(context.Background(), projectid).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.ProjectsDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

}

func cleanupTerraformFiles(modulePath string) error {

	// List of files/directories to remove
	tfFiles := []string{
		".terraform",
		".terraform.lock.hcl",
		"terraform.tfstate",
		"terraform.tfstate.backup",
	}

	for _, item := range tfFiles {
		fullPath := filepath.Join(modulePath, item)
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", fullPath, err)
		}
	}

	return nil
}

func randomSecretName() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 4)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("terra%s", string(b))
}
