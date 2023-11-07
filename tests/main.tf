terraform {
  required_providers {
    qernal = {
      source = "registry.terraform.io/hashicorp/qernal"
    }
  }
  required_version = ">= 1.1.0"
}

provider "qernal" {
  host_hydra = "https://hydra.qernal-bld.dev"
  host_chaos = "https://chaos.qernal-bld.dev"
}

resource "qernal_organisation" "organisation_1" {
  name = "organisation_1_name"
}

resource "qernal_project" "project_1" {
  name   = "project_1_name_updated_2"
  org_id = qernal_organisation.organisation_1.id
}

resource "qernal_project" "project_2" {
  name   = "project_2_name"
  org_id = qernal_organisation.organisation_1.id
}

resource "qernal_secret" "environment_secret" {
  project_id = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name = "ENVIRONMENT_SECRET_NAME_111"
  type = "environment"
  payload = {
    environment_value = "MY_FIRST_SECRET_11"
  }
  encryption = "keys/dek/0"
}

resource "qernal_secret" "registry_secret" {
  project_id = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name = "REGISTRY_SECRET_NAME_1"
  type = "registry"
  payload = {
    registry = ""
    registry_value = ""
  }
  encryption = "keys/dek/0"
}

resource "qernal_secret" "certificate_secret" {
  project_id = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name = "CERTIFICATE_SECRET_NAME"
  type = "certificate"
  payload = {
    certificate = ""
    certificate_value = ""
  }
  encryption = "keys/dek/0"
}

output "qernal_organisation" {
  value = qernal_organisation.organisation_1
}

output "qernal_project" {
  value = qernal_project.project_1
}
