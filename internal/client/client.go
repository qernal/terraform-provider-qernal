package client

import (
	"context"
	"encoding/json"
	"fmt"
	openapiclient "github.com/qernal/openapi-chaos-go-client"
	"io"
	"net/http"
	"terraform-provider-qernal/pkg/oauth"
)

type QernalAPIClient struct {
	openapiclient.APIClient
}

func New(ctx context.Context, hostHydra, hostChaos, token string) (client QernalAPIClient, err error) {

	oauthClient := oauth.NewOauthClient(hostHydra)
	err = oauthClient.ExtractClientIDAndClientSecretFromToken(token)
	if err != nil {
		return
	}

	accessToken, err := oauthClient.GetAccessTokenWithClientCredentials()
	if err != nil {
		return
	}

	configuration := &openapiclient.Configuration{
		Servers: openapiclient.ServerConfigurations{
			{
				URL: hostChaos + "/v1",
			},
		},
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		},
	}
	apiClient := openapiclient.NewAPIClient(configuration)

	return QernalAPIClient{
		APIClient: *apiClient,
	}, nil
}

func ParseResponseData(res *http.Response) (resData interface{}, err error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return
	}
	return data, nil
}

type ResponseData struct {
	Data string `json:"data"`
}
