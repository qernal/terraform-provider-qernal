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
  name   = "project_1_name_updated"
  org_id = qernal_organisation.organisation_1.id
}

output "qernal_organisation" {
  value = qernal_organisation.organisation_1
}

output "qernal_project" {
  value = qernal_project.project_1
}
