// Create a gateway policy with a set of CIDRs.
resource "gamefabric_protection_gatewaypolicy" "game_backend" {
  name         = "game-backend"
  display_name = "Game Backend"
  description  = "The game backend that game servers connect to"
  destination_cidrs = [
    # Azure Service
    "4.232.24.16/30",
    # Google Cloud Service
    "23.236.48.0/20",
    # Single IP
    "23.236.48.231/32",
  ]
}
