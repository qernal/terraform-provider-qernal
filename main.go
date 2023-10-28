package main

import (
	"fmt"
	"github.com/spf13/viper"
	"qernal-terraform-provider/pkg/oauth"
	"strings"
)

func main() {

	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	// TODO: will be replaced with the production token url of OAuth Server (Ory Hydra)
	tokenURL := "https://hydra.qernal-bld.dev/oauth2/token"

	qernalTokenEnv := viper.Get("QERNAL_TOKEN")
	if qernalTokenEnv == nil {
		panic("The qernal token is required")
	}

	qernalToken := qernalTokenEnv.(string)
	if !strings.Contains(qernalToken, "@") {
		panic("The qernal token is invalid")
	}

	clientID := strings.Split(qernalToken, "@")[0]
	clientSecret := strings.Split(qernalToken, "@")[1]

	oauthClient := oauth.NewOauthClient(tokenURL, clientID, clientSecret)
	token, err := oauthClient.GetAccessTokenWithClientCredentials()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(token)
}
