resource "qernal_secret_environment" "example" {
  name       = var.secret_name
  project_id = var.project_id
  value      = var.secret_value
}

output "secret_name" {
  value = qernal_secret_environment.example.name
}

output "secret_value" {
  value = qernal_secret_environment.example.value
}
