resource "gamefabric_armada" "this" {
  name        = "myarmada"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name
  replicas = [
    {
      region_type = "baremetal"
      # min_replicas = 0: buffer_size is the effective minimum.
      # The autoscaler targets buffer_size Ready servers and
      # scales down freely when demand drops. Recommended when
      # dynamic_buffer is enabled.
      min_replicas = 0
      max_replicas = 200
      buffer_size  = 10
    },
    {
      region_type = "cloud"
      # min_replicas > 0: static floor, keeps at least this many
      # servers running at all times. Useful for pre-warming ahead
      # of a launch. Must be >= buffer_size.
      min_replicas = 5
      max_replicas = 100
      buffer_size  = 5
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
      envs = [
        {
          name  = "REGION"
          value = data.gamefabric_region.europe.name
        },
        {
          name = "REGION_TYPE"
          value_from = {
            field_path = "metadata.regionTypeName"
          }
        }
      ]
      ports = [
        {
          name           = "game"
          protocol       = "TCP"
          container_port = 27015
          policy         = "Dynamic"
        }
      ]
    }
  ]
}

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
