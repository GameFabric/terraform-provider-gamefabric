# Get all token services without any filtering.
data "gamefabric_steelshield_tokenservices" "all" {}

# Get token services filtered by labels.
data "gamefabric_steelshield_tokenservices" "platforms_team" {
  label_filter = {
    team = "platforms"
  }
}
