resource "gamefabric_imageupdater" "name" {
  branch = data.gamefabric_branch.prod.name
  image  = data.gamefabric_image.gameserver.image
  target = gamefabric_armada.this.imageupdater_target
}


data "gamefabric_branch" "prod" {
  display_name = "prod"
}

data "gamefabric_image" "gameserver" {
  branch = data.gamefabric_branch.prod.name
  image  = "gameserver"
  tag    = "latest"
}

resource "gamefabric_armada" "this" {
  lifecycle {
    ignore_changes = [
      containers[0].image, # Prevents terraform diffs when image is updated by the image updater
    ]
  }
  name        = "myarmada"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name
  replicas = [
    {
      region_type  = "baremetal"
      min_replicas = 10
      max_replicas = 200
      buffer_size  = 10
    }
  ]
  containers = [
    {
      name  = "default" # the gameserver container should always be named "default"
      image = data.gamefabric_image.gameserver.image_target
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
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
