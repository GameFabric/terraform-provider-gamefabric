resource "gamefabric_armadaset" "this" {
  name        = "myarmadaset"
  environment = data.gamefabric_environment.prod.name

  regions = [
    {
      name = "europe"
      replicas = [
        {
          region_type = "baremetal"
          # min_replicas = 0: buffer_size is the effective minimum.
          # Autoscaler targets buffer_size Ready servers and scales
          # down freely when demand drops. Recommended when
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
      envs = [
        {
          name  = "REGION"
          value = "europe"
        }
      ]
    },
    {
      name = "us-east"
      replicas = [
        {
          region_type  = "baremetal"
          min_replicas = 0
          max_replicas = 200
          buffer_size  = 10
        },
        {
          region_type  = "cloud"
          min_replicas = 5
          max_replicas = 100
          buffer_size  = 5
        }
      ]
      envs = [
        {
          name  = "REGION"
          value = "us-east"
        }
      ]
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
          name  = "COMMON_ENV"
          value = "this-is-the-same-for-all-regions"
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
