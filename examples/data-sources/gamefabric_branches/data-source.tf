# Get all branches without filtering.
data "gamefabric_branches" "all" {
}

# Get branches filtered by labels.
data "gamefabric_branches" "production_platform" {
  label_filter = {
    env  = "prod"
    team = "platform"
  }
}
