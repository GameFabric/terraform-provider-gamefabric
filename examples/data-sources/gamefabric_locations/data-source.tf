# Get all locations without filtering.
# This returns all available locations in the system.
data "gamefabric_locations" "all" {
}

# Filter locations by provider type.
# Common types include "cloud", "baremetal", "edge", etc.
data "gamefabric_locations" "cloud" {
  type = "cloud"
}

# Filter locations by geographic attributes.
# You can filter by continent, country, or city.
data "gamefabric_locations" "europe" {
  continent = "europe"
}

# Combine multiple filters.
# This example gets all baremetal locations in Europe.
data "gamefabric_locations" "eu_baremetal" {
  type      = "baremetal"
  continent = "europe"
}

# Filter by city for specific location granularity.
data "gamefabric_locations" "frankfurt" {
  city = "frankfurt"
}

# Filter by country to get all locations in a specific country.
data "gamefabric_locations" "germany" {
  country = "germany"
}

# Use regular expression to filter location names.
# This is useful for pattern matching location names.
# For example, to get all GCP locations:
data "gamefabric_locations" "gcp" {
  name_regex = ".*-gcp-.*"
}

# Combine regex with other filters.
# Get all AWS locations in North America.
data "gamefabric_locations" "na_aws" {
  continent  = "north-america"
  name_regex = ".*-aws-.*"
}

# Access the list of location names.
# The data source returns a 'names' attribute containing all matching location names.
output "available_cloud_locations" {
  description = "List of all cloud location names"
  value       = data.gamefabric_locations.cloud.names
}

# Use locations in a region resource.
# This is a common pattern for dynamically selecting locations.
data "gamefabric_environment" "prod" {
  display_name = "Production"
}

resource "gamefabric_region" "europe" {
  name         = "europe"
  display_name = "Europe"
  environment  = data.gamefabric_environment.prod.name

  types = {
    baremetal = {
      # Use all baremetal locations in Europe
      locations  = data.gamefabric_locations.eu_baremetal.names
      scheduling = "Packed"
      env = [
        {
          name  = "GAME_REGION"
          value = "EU"
        }
      ]
    }
  }
}
