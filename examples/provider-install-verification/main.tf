terraform {
  required_providers {
    qernal = {
      source = "registry.terraform.io/hashicorp/qernal"
    }
  }
}

provider "qernal" {
  token = "client_id@client_secret"
  host = "https://api.staging.qernal.com/v1"
}

data "qernal_coffees" "example" {}
