# Get a volumestore by its unique name.
data "gamefabric_volumestore" "by_name" {
  name = "storage-eu"
}

# Get a volumestore by region.
data "gamefabric_volumestore" "by_region" {
  region = "eu"
}

