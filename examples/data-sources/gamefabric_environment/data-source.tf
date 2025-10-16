# Get an environment by its unique name.
#
# The name must match exactly and is case-sensitive.
data "gamefabric_environment" "dev" {
  name = "dev" # The name is the technical identifier (limited to 4 characters).
}
# output = "gamefabric_environment_v1.test.name"

# Get a branch by display name.
#
# Display names are user-friendly names that can contain spaces and special characters.
#
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple branches share the same display name, an error will occur.
data "gamefabric_environment" "prod" {
  display_name = "Production Environment"
}
