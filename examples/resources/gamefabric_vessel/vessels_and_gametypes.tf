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

  # Create multiple vessels per game mode
  vessels = flatten([
    for mode_key, mode in local.game_modes : [
      for i in range(3) : {
        key          = "${mode_key}-${i}"
        mode_key     = mode_key
        display_name = mode.display_name
        mode_value   = mode.mode_value
        index        = i
      }
    ]
  ])
}

resource "gamefabric_vessel" "this" {
  for_each    = { for v in local.vessels : v.key => v }
  name        = "${each.value.mode_key}-${each.value.index}"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name

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
          name  = "SERVER_NAME"
          value = "${each.value.display_name} ${each.value.index + 1} - ${data.gamefabric_region.europe.display_name}"
        },
        {
          name  = "GAME_MODE"
          value = each.value.mode_value
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
