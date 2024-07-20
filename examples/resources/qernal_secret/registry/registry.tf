
terraform {
  required_providers {
    qernal = {
      source = "qernal/qernal"
    }
  }
}



variable "qernal_token" {}
variable "project_id" {}

variable "auth_token" {}


provider "qernal" {
  token      = var.qernal_token
}

resource "qernal_secret_registry" "example" {
  name         = "my-registry"
  project_id   = "project-123"
  registry_url = "ghcr.io"
  auth_token   = "your-auth-token-here"
}

# Output the created_at date of the secret registry
output "registry_created_at" {
  value = qernal_secret_registry.example.date.created_at
}

# Output the revision number of the secret registry
output "registry_revision" {
  value = qernal_secret_registry.example.revision
}