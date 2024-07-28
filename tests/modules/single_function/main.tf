data "qernal_provider" "digitalocean" {
  name = "digitalocean"
}

resource "qernal_function" "function" {
  project_id  = var.project_id
  version     = "1.0.0"
  name        = var.function_name
  description = "Hello world"
  image       = "ealen/echo-server:0.9.2"
  port        = 80
  type        = "http"

  scaling = {
    high = 80
    low  = 20
    type = "cpu"
  }

  size = {
    cpu    = 128
    memory = 128
  }
  compliance = []

  deployment {
    location = {
      provider_id = data.qernal_provider.digitalocean.id
      country     = "GB"
    }

    replicas = {
      min = 1
      max = 1
      affinity = {
        cloud   = false
        cluster = false
      }
    }
  }

  route {
    path    = "/*"
    methods = "GET, HEAD"
  }
}
