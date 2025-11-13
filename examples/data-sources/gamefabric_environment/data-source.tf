# Get an environment by its unique name.
data "gamefabric_environment" "dev" {
  name = "dev"
}

# Get an environment by display name.
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple environments share the same display name, an error will occur.
data "gamefabric_environment" "prod" {
  display_name = "Production Environment"
}
