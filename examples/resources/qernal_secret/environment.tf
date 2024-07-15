resource "qernal_secret_environment" "example" {
  name       = "PORT"
  project_id = "project-123"
  value      = "8080"
}

# Output the created_at date of the secret environment
output "env_created_at" {
  value = qernal_secret_environment.example.date.created_at
}

# Output the revision number of the secret environment
output "env_revision" {
  value = qernal_secret_environment.example.revision
}