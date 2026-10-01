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

# Armada with a Prometheus metrics port.
#
# The "Metric" policy is a GameFabric-native abstraction: it maps to the Agones
# "None" port policy (no public host port) and automatically injects the
# g8c.io/gameserver-scrape label and g8c.io/metrics-endpoints annotation onto
# each GameServer pod, enabling Prometheus scraping without any manual
# label/annotation management.
resource "gamefabric_armada" "this" {
  name        = "myarmada"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name
  replicas = [
    {
      region_type  = "baremetal"
      min_replicas = 10
      max_replicas = 200
      buffer_size  = 10
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
