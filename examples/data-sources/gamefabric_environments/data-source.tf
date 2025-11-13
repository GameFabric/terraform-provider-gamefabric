# Get all environments without any filtering.
data "gamefabric_environments" "all" {
}

# Get environments filtered by labels.
data "gamefabric_environments" "dev_stage" {
  label_filter = {
    stage = "dev"
  }
}
