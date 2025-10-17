// Create a gateway policy with a set of Classless Inter-Domain Routing (CIDR).
resource "gamefabric_protection_gatewaypolicy" "game_backend" {
  name = "game-backend"
  display_name = "Game Backend"
  description = "The game backend that game servers connect to"
  destination_cidrs = [
    # Data center 1
    "172.123.64.201/32",
    "173.124.0.0/16",

    # Data center 2
    "155.12.91.8/32"
  ]
}
