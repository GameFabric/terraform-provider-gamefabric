data "gamefabric_volumestore" "eu" {
  name = "storage-eu"
}

data "gamefabric_environment" "prod" {
  name = "prod"
}

data "gamefabric_branch" "prod" {
  display_name = "prod"
}

data "gamefabric_image" "gameserver" {
  branch = data.gamefabric_branch.prod.name
  image  = "gameserver"
  tag    = "1.2.3"
}

data "gamefabric_region" "europe" {
  name        = "europe"
  environment = data.gamefabric_environment.prod.name
}

# Setup a volume to be used by a vessel
resource "gamefabric_volume" "example" {
  name         = "example-volume"
  environment  = data.gamefabric_environment.prod.name
  volume_store = data.gamefabric_volumestore.eu.name

  capacity = "100M"
}

resource "gamefabric_vessel" "this" {
  name        = "my-vessel"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name

  containers = [
    {
      name      = "default" # the game server container should always be named "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      volume_mounts = [
        {
          name       = gamefabric_volume.example.name
          mount_path = "/data"
        }
      ]
    }
  ]
  volumes = [
    {
      name = gamefabric_volume.example.name
      persistent = {
        volume_name = gamefabric_volume.example.name
      }
    }
  ]
}
