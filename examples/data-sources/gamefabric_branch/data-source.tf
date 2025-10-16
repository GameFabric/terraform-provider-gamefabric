# Get a branch by its unique name.
data "gamefabric_branch" "main" {
  name = "main"
}

# Get a branch by display name.
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple branches share the same display name, an error will occur.
data "gamefabric_branch" "production" {
  display_name = "Production Branch"
}
