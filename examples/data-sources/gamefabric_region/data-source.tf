# Get a specific regions by its unique name.
data "gamefabric_region" "europe" {
  name        = "eu"
  environment = "dev" # The environment this region belongs to (required)
}

# Get a region by its display name.
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple branches share the same display name, an error will occur.
data "gamefabric_region" "europe" {
  display_name = "Europe"
  environment  = "dev" # The environment this region belongs to (required)
}
