# This example demonstrates how to create a GameFabric Formation with persistent volumes.
# Some attributes have been omitted for brevity.
resource "gamefabric_formation" "formation_with_volumes" {
  name        = "formation-with-volumes"
  environment = data.gamefabric_environment.prod.name

  volume_templates = [
    {
      name              = "persist-save-files"
      volume_store_name = "eu" # Required to exist.
      capacity          = "1Gi"
    }
  ]

  vessels = [
    {
      name        = "vessel-1"
      region      = data.gamefabric_region.europe.name
    }
  ]

  containers = [
    {
      name = "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      ports = [
        {
          name                = "game"
          protocol            = "UDP"
          policy              = "Passthrough"
        }
      ]
      volume_mounts = [
        {
          name          = "mounted-save-files"
          mount_path    = "/data/save_files"
        }
      ]
    }
  ]

  // Volumes
  volumes = [
    {
      name = "mounted-save-files"
      persistent = {
        volume_name = "persist-save-files"
      }
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
