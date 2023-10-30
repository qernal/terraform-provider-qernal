terraform {
  required_providers {
    qernal = {
      source  = "registry.terraform.io/hashicorp/qernal"
    }
  }
  required_version = ">= 1.1.0"
}

provider "qernal" {
  host_hydra = "https://hydra.qernal-bld.dev/oauth2/token"
  host_chaos = "https://chaos.qernal-bld.dev/v1"
  token = "client_id@client_secret"
}

resource "qernal_organisation" "organisation1111" {
  name = "sdfsfsdfds"
}

output "qernal_organisation" {
  value = qernal_organisation.organisation1111
}
