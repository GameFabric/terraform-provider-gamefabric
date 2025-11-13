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
      name      = "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
    }
  ]
}
