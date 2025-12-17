# Get all secrets without any filtering.
data "gamefabric_secrets" "all_of_environment" {
  environment = gamefabric_environment.prod.name
}

# Get secrets filtered by labels.
data "gamefabric_secrets" "game_type_alpha_configurations" {
  environment = gamefabric_environment.prod.name
  label_filter = {
    game-type = "alpha"
  }
}
