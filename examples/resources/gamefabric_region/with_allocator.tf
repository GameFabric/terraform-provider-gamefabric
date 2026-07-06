data "gamefabric_allocator" "eu" {
  name = "eu-allocator"
}

data "gamefabric_environment" "prod" {
  name = "prod"
}

data "gamefabric_locations" "cloud_europe" {
  type      = "cloud"
  continent = "europe"
}

resource "gamefabric_region" "europe" {
  name         = "europe"
  display_name = "Europe"
  environment  = data.gamefabric_environment.prod.name
  allocator    = data.gamefabric_allocator.eu.name

  types = [
    {
      name      = "cloud"
      locations = data.gamefabric_locations.cloud_europe.names
    }
  ]
}
