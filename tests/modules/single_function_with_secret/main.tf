data "qernal_provider" "digitalocean" {
  name = "digitalocean"
}

resource "qernal_function" "function" {
  project_id  = var.project_id
  version     = "1.0.0"
  name        = var.function_name
  description = "Hello world"
  image       = "testcontainers/helloworld:1.1.0"
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
  compliance = ["ipv6"] // TODO: allow this to be blank

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
    methods = ["GET", "HEAD"]
    weight  = 100
  }

  secret {
    name      = "ENV_VAR"
    reference = var.secret_reference
  }

}
