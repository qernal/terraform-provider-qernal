resource "qernal_secret_certificate" "example" {
  name       = "my-certificate"
  project_id = "project-123"

  certificate = <<EOT
-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUJgLsHCx/VLs0m6NkVXTKgzRqAQMwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMzA3MTQxNTMwMzhaFw0yNDA3
MTMxNTMwMzhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQC5p0Hn9WuPHuKTGK2IpOInPbOmZ3hPNmcj6SWDTa4s
...
-----END CERTIFICATE-----
EOT

  certificate_value = <<EOT
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC5p0Hn9WuPHuKT
GK2IpOInPbOmZ3hPNmcj6SWDTa4sWfV2BHHWxKaJCqKP8wGjIK5Xzim7xBxp+HKS
fWmUFiRKh7kjgXUtqX3WfaZLVcCLgOK3OLwT5pZ/DcUkDKv7+JcLWt1vFxJcSzCk
...
-----END PRIVATE KEY-----
EOT
}

# Output the created_at date of the secret certificate
output "certificate_created_at" {
  value = qernal_secret_certificate.example.date.created_at
}

# Output the revision number of the secret certificate
output "certificate_revision" {
  value = qernal_secret_certificate.example.revision
}