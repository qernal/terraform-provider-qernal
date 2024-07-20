package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestCertificateSecret(t *testing.T) {
	t.Parallel()

	envs := checkRequiredEnvs(t, []string{"QERNAL_TOKEN", "QERNAL_PROJECT"})

	qernalToken := envs["QERNAL_TOKEN"]
	projectID := envs["QERNAL_PROJECT"]

	publicKey, privateKey := generateCertificate(t)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../resources/qernal_secret/certificate",
		Vars: map[string]interface{}{
			"project_id":   projectID,
			"qernal_token": qernalToken,
			"public_key":   publicKey,
			"private_key":  privateKey,
		},
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Test outputs
	certRevision := terraform.Output(t, terraformOptions, "certificate_revision")

	// Verify that the revision is a number
	assert.Regexp(t, "^[0-9]+$", certRevision, "certificate_revision should be a number")

	// Verify the name of the created resource
	name := terraform.Output(t, terraformOptions, "qernal_secret_certificate_name")
	assert.Equal(t, "my-certificate", name)

	// Verify the reference format
	// TODO: uncomment when new release is made
	// reference := terraform.Output(t, terraformOptions, "qernal_secret_certificate_reference")
	// expectedReference := fmt.Sprintf("projects:%s/my-certificate", projectID)
	// assert.Equal(t, expectedReference, reference, "Reference should be in the format projects:PROJECT_ID/SECRET_NAME")
}

func generateCertificate(t *testing.T) (string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180), // Valid for 180 days
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return string(publicKeyPEM), string(privateKeyPEM)
}
