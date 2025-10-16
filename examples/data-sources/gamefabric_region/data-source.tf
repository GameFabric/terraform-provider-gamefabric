# First, get the environment that contains the region.
# Regions are scoped to specific environments.
data "gamefabric_environment" "dev" {
  display_name = "Development"
}

# Get a region by its display name.
# The display name is the user-friendly name.
# Both display_name and environment are required when using display_name.
# Note: Either 'name' or 'display_name' can be used, but not both.
data "gamefabric_region" "europe_by_display_name" {
  display_name = "Europe"
  environment  = data.gamefabric_environment.dev.name
}

# Alternatively, get a region by its unique name.
# The name is the technical identifier.
data "gamefabric_region" "europe_by_name" {
  name        = "europe"
  environment = data.gamefabric_environment.dev.name
}

# Access region attributes after retrieval.
# The region data source returns detailed information about infrastructure types and locations.
output "region_info" {
  description = "Information about the Europe region"
  value = {
    name         = data.gamefabric_region.europe_by_display_name.name
    display_name = data.gamefabric_region.europe_by_display_name.display_name
    labels       = data.gamefabric_region.europe_by_display_name.labels
    types        = data.gamefabric_region.europe_by_display_name.types
  }
}
