data "gamefabric_image" "gameserver" {
  branch = data.gamefabric_branch.example.name
  image  = "gameserver"
  tag    = "latest"
}

resource "gamefabric_armada" "this" {
  name        = "myarmada"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name
  replicas = [
    # ...
  ]
  containers = [
    {
      name  = "default" # The gameserver container should always be named "default".
      image = data.gamefabric_image.gameserver.image_target
    }
  ]
}
