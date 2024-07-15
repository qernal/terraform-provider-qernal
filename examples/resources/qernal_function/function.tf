# Data source for the qernal provider
data "qernal_provider" "example" {
  name = "AWS"
}

resource "qernal_function" "example" {
  name        = "my-function"
  project_id  = "project-123"
  description = "An example function"
  type        = "HTTP"
  version     = "1.0.0"
  image       = "docker.io/myrepo/myimage:latest"
  port        = 8080
  compliance  = ["soc", "ipv6"]

  size {
    cpu    = 1
    memory = 512
  }

  scaling {
    type = "CPU"
    low  = 20
    high = 80
  }

  deployments {
    location {
      continent   = data.qernal_provider.example.continents[0]
      country     = data.qernal_provider.example.countries[0]
      city        = data.qernal_provider.example.cities[0]
      provider_id = data.qernal_provider.example.id
    }
    replicas {
      min = 1
      max = 5
      affinity {
        cloud   = true
        cluster = false
      }
    }
  }

  routes {
    path    = "/api"
    methods = ["GET", "POST"]
    weight  = 100
  }

  secrets {
    name      = "API_KEY"
    reference = "secret_ref_123"
  }
}

# Output the function ID
output "function_id" {
  value = qernal_function.example.id
}