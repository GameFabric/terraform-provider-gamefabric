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
  types = [
    {
      name      = "baremetal"
      locations = data.gamefabric_locations.baremetal_europe.names
    },
    {
      name      = "gcp"
      locations = data.gamefabric_locations.gcp.names
    },
    {
      name      = "aws"
      locations = data.gamefabric_locations.aws.names
    }
  ]
}
