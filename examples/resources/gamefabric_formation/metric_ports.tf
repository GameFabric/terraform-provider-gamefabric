data "gamefabric_branch" "prod" {
  display_name = "prod"
}

data "gamefabric_image" "gameserver" {
  branch = data.gamefabric_branch.prod.name
  image  = "gameserver"
  tag    = "1.2.3"
}

data "gamefabric_environment" "prod" {
  name = "prod"
}

data "gamefabric_region" "europe" {
  name        = "europe"
  environment = data.gamefabric_environment.prod.name
}

# Enables Prometheus scraping of the game server through a Metric port.
resource "gamefabric_formation" "this" {
  name        = "my-formation"
  environment = data.gamefabric_environment.prod.name

  vessels = [
    {
      name   = "vessel-1"
      region = data.gamefabric_region.europe.name
    }
  ]

  containers = [
    {
      name      = "default" # the game server container should always be named "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      ports = [
        {
          name           = "game"
          protocol       = "UDP"
          container_port = 7777
          policy         = "Dynamic"
        },
        {
          # Metric ports are scraped by Prometheus. protocol must be TCP.
          # path defaults to /metrics when omitted.
          name           = "metrics"
          protocol       = "TCP"
          container_port = 9090
          policy         = "Metric"
          path           = "/metrics"
        }
      ]
    }
  ]
}
