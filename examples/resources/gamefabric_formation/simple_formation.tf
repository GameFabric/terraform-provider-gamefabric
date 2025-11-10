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

resource "gamefabric_formation" "this" {
  name        = "my-formation"
  environment = data.gamefabric_environment.prod.name

  vessels = [
    {
      name = "vessel-1"
      region = data.gamefabric_region.europe.name
    }
  ]

  containers = [
    {
      name      = "default" # the gameserver container should always be named "default"
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
          value = "My Server ${data.gamefabric_region.europe.display_name}"
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
