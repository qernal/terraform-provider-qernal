# Data source for the qernal provider
data "qernal_provider" "example" {
  name = "GCP"
}

resource "qernal_function" "example" {
  name        = "my-function"
  project_id  = "a8d86e21-df21-407c-854b-f0aecec00042"
  description = "An example function"
  type        = "http"
  version     = "1.0.0"
  image       = "docker.io/myrepo/myimage:latest"
  port        = 8080
  compliance  = ["soc2"]

  size = {
    cpu    = 1
    memory = 512
  }

  scaling = {
    type = "cpu"
    low  = 20
    high = 80
  }

  deployment {
    location = {
      continent   = data.qernal_provider.example.continents[0]
      country     = data.qernal_provider.example.countries[1]
      city        = data.qernal_provider.example.cities[0]
      provider_id = data.qernal_provider.example.id
    }
    replicas = {
      min = 1
      max = 5
      affinity = {
        cloud   = true
        cluster = false
      }
    }
  }

  route {
    path    = "/api"
    methods = ["GET", "POST"]
    weight  = 80
  }

  # secrets = [{
  #   name      = "API_KEY"
  #   reference = "projects:a8d86e21-df21-407c-854b-f0aecec0004/GHCR"
  # }]
}


# Output the function ID
output "function_id" {
  value = qernal_function.example.id
}
