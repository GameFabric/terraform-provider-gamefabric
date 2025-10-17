data "gamefabric_locations" "gcp" {
  type       = "cloud"
  continent  = "europe"
  name_regex = ".*-gcp"
}

data "gamefabric_locations" "aws" {
  type       = "cloud"
  continent  = "europe"
  name_regex = ".*-aws"
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
    "gcp" = {
      locations = data.gamefabric_locations.gcp.names
    },
    "aws" = {
      locations = data.gamefabric_locations.aws.names
    }
  }
}
