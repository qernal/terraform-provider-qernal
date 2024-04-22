package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"terraform-provider-qernal/pkg/oauth"

	openapiclient "github.com/qernal/openapi-chaos-go-client"
	"golang.org/x/crypto/nacl/box"
)

type QernalAPIClient struct {
	openapiclient.APIClient
}

func New(ctx context.Context, hostHydra, hostChaos, token string) (client QernalAPIClient, err error) {

	oauthClient := oauth.NewOauthClient(hostHydra)
	err = oauthClient.ExtractClientIDAndClientSecretFromToken(token)
	if err != nil {
		return QernalAPIClient{}, err
	}

	accessToken, err := oauthClient.GetAccessTokenWithClientCredentials()
	if err != nil {
		return QernalAPIClient{}, err
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

func EncryptLocalSecret(pk, secret string) (string, error) {
	secretBytes := []byte(secret)
	pubKey, err := base64.StdEncoding.DecodeString(pk)
	if err != nil {
		return "", err
	}

	// Create a slice with enough capacity for both secret and public key
	privateKey := make([]byte, 0, len(secretBytes)+len(pubKey))
	privateKey = append(privateKey, secretBytes...)
	privateKey = append(privateKey, pubKey...)
	plaintextBytes := []byte(secret)

	var privateKeyArray [32]byte
	copy(privateKeyArray[:], privateKey)

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", err
	}

	encrypted := box.Seal(nonce[:], plaintextBytes, &nonce, &privateKeyArray, new([32]byte))

	return base64.StdEncoding.EncodeToString(encrypted), nil
}
