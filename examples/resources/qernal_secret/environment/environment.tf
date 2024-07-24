
terraform {
  required_providers {
    qernal = {
      source = "qernal/qernal"
    }
  }
}


variable "qernal_token" {}


provider "qernal" {
  token = var.qernal_token
}


variable "project_id" {}

resource "qernal_secret_environment" "example" {
  name       = "PORT"
  project_id = var.project_id
  value      = "8080"
}

output "qernal_secret_environment_name" {
  value = qernal_secret_environment.example.name
}

output "qernal_secret_environment_value" {
  value = qernal_secret_environment.example.value
}


# Output the created_at date of the secret registry
output "env_revision" {
  value = qernal_secret_environment.example.revision
}


// TODO: uncomment when new release is made
# output "qernal_secret_environment_reference" {
#   value = qernal_secret_environment.example.reference
# }
