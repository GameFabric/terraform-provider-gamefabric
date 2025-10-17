# Get all regions without any filtering.
data "gamefabric_regions" "all" {
  environment = "dev" # The environment this region belongs to (required)
}

# Get regions filtered by labels.
# Note:
# - Only regions that have ALL specified labels with exact matching values are returned.
data "gamefabric_regions" "eu" {
  environment = "dev" # The environment this region belongs to (required)
  label_filter = {
    env = "prod"
  }
}
