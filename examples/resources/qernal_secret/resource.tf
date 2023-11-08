# Manage example qernal_secret.
resource "qernal_secret" "environment_secret" {
  project_id = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name       = "ENVIRONMENT_SECRET_NAME_1"
  type       = "environment"
  payload = {
    environment_value = "MY_FIRST_SECRET"
  }
}

resource "qernal_secret" "registry_secret" {
  project_id = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name       = "REGISTRY_SECRET_NAME_1"
  type       = "registry"
  payload = {
    registry       = ""
    registry_value = ""
  }
}

resource "qernal_secret" "certificate_secret" {
  project_id = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name       = "CERTIFICATE_SECRET_NAME"
  type       = "certificate"
  payload = {
    certificate       = ""
    certificate_value = ""
  }
}