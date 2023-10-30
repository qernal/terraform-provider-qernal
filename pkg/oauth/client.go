package oauth

import (
	"context"
	"errors"
	"golang.org/x/oauth2/clientcredentials"
	"strings"
)

type OAuthClient interface {
	GetAccessTokenWithClientCredentials() (string, error)
	ExtractClientIDAndClientSecretFromToken(string) error
}

type oauthClient struct {
	tokenURL     string
	clientID     string
	clientSecret string
}

func (oc *oauthClient) GetAccessTokenWithClientCredentials() (token string, err error) {
	config := clientcredentials.Config{
		ClientID:     oc.clientID,
		ClientSecret: oc.clientSecret,
		TokenURL:     oc.tokenURL,
	}

	oauthToken, err := config.Token(context.TODO())
	if err != nil {
		return
	}
	return oauthToken.AccessToken, nil

}

func NewOauthClient(tokenURL string) OAuthClient {
	return &oauthClient{
		tokenURL: tokenURL,
	}
}

func (oc *oauthClient) ExtractClientIDAndClientSecretFromToken(token string) (err error) {
	if !strings.Contains(token, "@") || strings.Count(token, "@") > 1 {
		err = errors.New("the qernal token is invalid")
		return
	}

	oc.clientID = strings.Split(token, "@")[0]
	oc.clientSecret = strings.Split(token, "@")[1]

	return nil
}
