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

resource "gamefabric_armadaset" "this" {
  name        = "myarmadaset"
  environment = data.gamefabric_environment.prod.name

  regions = [
    {
      name = "europe"
      replicas = [
        {
          region_type  = "baremetal"
          min_replicas = 1
          max_replicas = 2
          buffer_size  = 1
        },
        {
          region_type  = "cloud"
          min_replicas = 1
          max_replicas = 2
          buffer_size  = 1
        }
      ]
      envs = [
        {
          name  = "REGION"
          value = "europe"
        },
        {
          name = "BACKEND_CONFIG"
          value_from = {
            config_file = "example-eu-config.ini"
          }
        },
        {
          name = "EXAMPLE_POD_FIELD"
          value_from = {
            field_path = "metadata.name"
          }
        }
      ]
      labels = {
        region = "europe"
      }
    },
    {
      name = "us-east"
      replicas = [
        {
          region_type  = "baremetal"
          min_replicas = 1
          max_replicas = 2
          buffer_size  = 1
        },
        {
          region_type  = "cloud"
          min_replicas = 1
          max_replicas = 2
          buffer_size  = 1
        }
      ]
      envs = [
        {
          name  = "REGION"
          value = "us-east"
        },
        {
          name = "BACKEND_CONFIG"
          value_from = {
            config_file = "example-us-config.ini"
          }
        },
        {
          name = "EXAMPLE_POD_FIELD"
          value_from = {
            field_path = "metadata.name" # can we somehow define all the allowed keys here? Does this make sense?
          }
        }
      ]
      labels = {
        region = "us-east"
      }
    }
  ]


  gameserver_labels = {
    example = "value"
  }

  containers = [
    {
      name  = "default" # the gameserver container should always be named "default"
      image = data.gamefabric_image.gameserver.image_target
      # args    = ["example-arg"]
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
        },
      ]
      ports = [
        {
          name           = "game"
          protocol       = "TCP"
          container_port = 27015
          policy         = "Dynamic"
        }
      ]
      config_files = [
        {
          name       = "example-config.ini"
          mount_path = "/config/example-config.ini"
        }
      ]
    }
  ]

  health_checks = {
    initial_delay_seconds = 180
    period_seconds        = 5
    failure_threshold     = 3
  }

  termination_configuration = {
    grace_period_seconds = 30
  }

  strategy = {
    rolling_update = {
      max_surge       = "25%"
      max_unavailable = "25%"
    }
  }
  profiling_enabled = false
}
