package tests

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/exp/rand"
)

func checkRequiredEnvs(t *testing.T, requiredEnvs []string, optionalEnvs ...string) map[string]string {
	envs := make(map[string]string)
	var missingEnvs []string

	// Check required envs
	for _, env := range requiredEnvs {
		value := os.Getenv(env)
		if value == "" {
			missingEnvs = append(missingEnvs, env)
		} else {
			envs[env] = value
		}
	}

	if len(missingEnvs) > 0 {
		t.Fatalf("Required environment variable(s) not set: %s", strings.Join(missingEnvs, ", "))
	}

	// Check optional envs
	for _, env := range optionalEnvs {
		value := os.Getenv(env)
		if value == "" {
			t.Logf("Optional environment variable not set: %s", env)
		} else {
			envs[env] = value
		}
	}

	return envs
}

func generateRandomToken() string {
	rand.Seed(299202)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, 8)
	for i := range token {
		token[i] = charset[rand.Intn(len(charset))]
	}
	return string(token)
}
