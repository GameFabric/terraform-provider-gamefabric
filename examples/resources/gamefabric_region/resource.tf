
# Lookup the locations for the baremetal type in Europe.
# This data source will return all locations that match the specified criteria.
data "gamefabric_locations" "eu_baremetal" {
  type      = "baremetal"
  continent = "europe"
  #city      = "frankfurt"
  #name_regex = ".*gcp"
}

# Get all environments as i want to setup the region the same for all environments.
data "gamefabric_environments" "all" {
}

# Create a region for Europe that uses the baremetal type.
resource "gamefabric_region_v1" "europe" {
  for_each     = data.gamefabric_environments.all.result
  name         = "europe"
  display_name = "Europe"
  labels = {
    continent = "europe"
  }

  environment = each.value.name
  types = {
    baremetal = {
      locations = data.gamefabric_locations_v1.eu_baremetal.names
      envs = [
        {
          name  = "GAME_REGION_NAME"
          value = "Europe"
        }
      ]
      scheduling = "Packed" # Options: "Packed", "Distributed"
    }
  }
}
