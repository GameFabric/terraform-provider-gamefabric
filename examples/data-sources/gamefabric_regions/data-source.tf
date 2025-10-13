data "gamefabric_environment" "dev" {
  display_name = "Development"
}

# Get all regions of the development environment.
data "gamefabric_regions" "all" {
  environment = data.gamefabric_environment.dev.name
}

# Get regions by labels
data "gamefabric_regions" "us" {
  environment = data.gamefabric_environment.dev.name
  label_filter = {
    continent = "us"
  }
}
