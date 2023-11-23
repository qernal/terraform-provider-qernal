package oauth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOauthClient_ExtractClientIDAndClientSecretFromToken(t *testing.T) {
	type testCase struct {
		name         string
		token        string
		expectError  bool
		clientID     string
		clientSecret string
	}

	tcs := []testCase{
		{
			name:        "invalid token",
			token:       "client_id@@client_secret",
			expectError: true,
		},
		{
			name:        "invalid token",
			token:       "client_id_client_secret",
			expectError: true,
		},
		{
			name:         "valid token",
			token:        "client_id@client_secret",
			expectError:  false,
			clientID:     "client_id",
			clientSecret: "client_secret",
		},
	}

	runInParallel := func(t *testing.T, tc testCase) {
		t.Parallel()
		oc := oauthClient{}
		err := oc.ExtractClientIDAndClientSecretFromToken(tc.token)
		if tc.expectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, oc.clientID, tc.clientID)
			assert.Equal(t, oc.clientID, tc.clientID)
		}
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runInParallel(t, tc)
		})
	}
}
