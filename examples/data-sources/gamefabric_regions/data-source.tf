# Get all regions without any filtering.
data "gamefabric_regions" "all" {
  environment = "dev" # The environment this region belongs to (required)
}

# Get regions filtered by labels.
data "gamefabric_regions" "eu" {
  environment = "dev" # The environment this region belongs to (required)
  label_filter = {
    env = "prod"
  }
}
