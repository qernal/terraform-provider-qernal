resource "qernal_secret_certificate" "example" {
  name              = var.secret_name
  project_id        = var.project_id
  certificate       = var.certificate
  certificate_value = var.cert_private_key
}


# Output
output "certificate_name" {
  value = qernal_secret_certificate.example.name
}
