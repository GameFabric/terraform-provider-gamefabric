# Get all branches without filtering.
data "gamefabric_branches" "all" {
}

# Get branches filtered by labels.
# Only branches that have ALL specified labels with exact matching values are returned.
data "gamefabric_branches" "production_platform" {
  label_filter = {
    env  = "prod"
    team = "platform"
  }
}
