package oauth

import (
	"context"
	"golang.org/x/oauth2/clientcredentials"
)

type OAuthClient interface {
	GetAccessTokenWithClientCredentials() (string, error)
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

func NewOauthClient(serverURL, clientID, clientSecret string) OAuthClient {
	return &oauthClient{
		tokenURL:     serverURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}
