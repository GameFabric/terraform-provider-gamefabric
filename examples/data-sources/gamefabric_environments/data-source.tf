# Get all environments without any filtering.
data "gamefabric_environments" "all" {
}

# Get environments filtered by labels.

# Note:
# - Only environments that have ALL specified labels with exact matching values are returned.
data "gamefabric_environments" "dev_stage" {
  label_filter = {
    environment_type = "non-prod"
    stage            = "dev"
  }
}
