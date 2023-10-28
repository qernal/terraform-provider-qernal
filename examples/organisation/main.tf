terraform {
  required_providers {
    qernal = {
      source  = "registry.terraform.io/hashicorp/qernal"
    }
  }
  required_version = ">= 1.1.0"
}

provider "qernal" {}

resource "qernal_organisation" "organisation1111" {
  name = "sdfsfsdfds"
}

output "edu_order" {
  value = qernal_organisation.organisation1111
}
