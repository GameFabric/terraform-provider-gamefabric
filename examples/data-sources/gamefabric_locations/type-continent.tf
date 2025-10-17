data "gamefabric_locations" "cloud_europe" {
  type      = "cloud"
  continent = "europe"
}

data "gamefabric_locations" "baremetal_europe" {
  type      = "baremetal"
  continent = "europe"
}

resource "gamefabric_region" "europe" {
  name         = "europe"
  display_name = "Europe"
  environment  = "prod"
  types = {
    "baremetal" = {
      locations = data.gamefabric_locations.baremetal_europe.names
    },
    "cloud" = {
      locations = data.gamefabric_locations.cloud_europe.names
    }
  }
}
