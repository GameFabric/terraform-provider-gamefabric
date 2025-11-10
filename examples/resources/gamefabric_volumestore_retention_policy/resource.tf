data "gamefabric_volumestore" "eu" {
  name = "storage-eu"
}

resource "gamefabric_volumestore_retention_policy" "name" {
  volume_store = data.gamefabric_volumestore.eu.name
  environment  = "prod"
  offline_snapshot = {
    count = 5
    days  = 7
  }
  online_snapshot = {
    count = 5
    days  = 7
  }
}
