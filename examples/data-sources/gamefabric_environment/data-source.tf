# Get an environment by its unique name.
data "gamefabric_environment" "dev" {
  name = "dev"
}

# Get a branch by display name.
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple branches share the same display name, an error will occur.
data "gamefabric_environment" "prod" {
  display_name = "Production Environment"
}
