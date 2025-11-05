# Get a volumestore by its unique name.
data "gamefabric_volumestore" "by_name" {
  name = "test-volumestore"
}

# Get a volumestore by region.
# Note:
# - Exactly one of `name` or `region` must be specified.
# - If multiple volumestores exist in the same region, the data source will return an error.
data "gamefabric_volumestore" "by_region" {
  region = "europe-west1"
}

