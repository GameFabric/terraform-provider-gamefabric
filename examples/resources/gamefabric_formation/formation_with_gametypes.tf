locals {
  game_modes = {
    survival = {
      display_name = "Survival"
      mode_value   = "survival"
    }
    creative = {
      display_name = "Creative"
      mode_value   = "creative"
    }
    hardcore = {
      display_name = "Hardcore"
      mode_value   = "hardcore"
    }
  }

  # Create vessels configuration for each game mode
  vessels_per_mode = {
    for mode_key, mode in local.game_modes : mode_key => [
      for i in range(3) : {
        name        = "${mode_key}-${i}"
        region      = data.gamefabric_region.europe.name
        description = "${mode.display_name} server ${i + 1}"
        suspend     = false
        override = {
          containers = [
            {
              envs = [
                {
                  name  = "SERVER_NAME"
                  value = "${mode.display_name} ${i + 1} - EU"
                }
              ]
            }
          ]
        }
      }
    ]
  }
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

resource "gamefabric_formation" "survival" {
  for_each    = local.game_modes
  name        = each.key
  environment = data.gamefabric_environment.prod.name

  vessels = local.vessels_per_mode[each.key]

  // Containers
  containers = [
    {
      name      = "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      envs = [
        {
          name  = "GAME_MODE"
          value = each.value.mode_value
        }
      ]
      ports = [
        {
          name           = "http"
          protocol       = "TCP"
          container_port = 27015
          policy         = "Dynamic"
        }
      ]
    }
  ]
}


