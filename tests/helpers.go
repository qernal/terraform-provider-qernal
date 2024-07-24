package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	openapi_chaos_client "github.com/qernal/openapi-chaos-go-client"
	"golang.org/x/oauth2/clientcredentials"
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

func deleteOrg(orgid string) {
	client := qernalClient()
	_, r, err := client.OrganisationsAPI.OrganisationsDelete(context.Background(), orgid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrganisationsAPI.OrganisationsDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
