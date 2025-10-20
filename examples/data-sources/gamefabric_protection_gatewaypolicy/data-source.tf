data "gamefabric_protection_gatewaypolicy" "game_backend_by_name" {
  name = "game-backend"
}

data "gamefabric_protection_gatewaypolicy" "game_backend_by_displayname" {
  display_name = "Game Backend"
}
