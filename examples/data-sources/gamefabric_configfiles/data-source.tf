# Get all config files without any filtering.
data "gamefabric_configfiles" "all_of_environment" {
  environment = gamefabric_environment.prod.name
}

# Get config files filtered by labels.

# Note:
# - Only config files that have ALL specified labels with exact matching values are returned.
data "gamefabric_configfiles" "game_type_alpha_configurations" {
  environment = gamefabric_environment.prod.name
  label_filter = {
    game-type = "alpha"
  }
}
