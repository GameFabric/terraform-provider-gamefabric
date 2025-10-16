# Get all environments without any filtering.
data "gamefabric_environments" "all" {
}

# Get environments filtered by labels.
#
# Labels are key-value pairs used to organize and categorize environments.
#
# Note:
# - Only environments that have ALL specified labels with exact matching values are returned.
data "gamefabric_environments" "non_prod" {
  label_filter = {
    environment_type = "non-prod"
  }
}
# output = array of environment objects

# Get environments with multiple label filters.
data "gamefabric_environments" "dev_stage" {
  label_filter = {
    environment_type = "non-prod"
    stage            = "dev"
  }
}
