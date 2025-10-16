# Get a specific regions by its unique name.
data "gamefabric_region" "europe" {
  name = "eu"
}

# Get a region by its display name.
#
# Display names are user-friendly names that can contain spaces and special characters.
#
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple branches share the same display name, an error will occur.
data "gamefabric_region" "europe" {
  display_name = "Europe"
}
