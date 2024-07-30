resource "qernal_secret_registry" "example" {
  name         = var.secret_name
  project_id   = var.project_id
  auth_token   = var.auth_token
  registry_url = var.registry_url
}


output "secret_name" {
  value = qernal_secret_registry.example.name
}

output "secret_value" {
  value = qernal_secret_registry.example.registry_url
}


