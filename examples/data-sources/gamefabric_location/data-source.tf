# Get a specific location by its unique name.
# This validates that the location exists and can be referenced.
# The name must match exactly and is case-sensitive.
data "gamefabric_location" "frankfurt" {
  name = "eu-central-1-fra"
}

# Use the location data source to validate availability.
# This is useful when you need to ensure a location is available
# before using it in region or other resource configurations.
data "gamefabric_location" "us_east" {
  name = "us-east-1-nyc"
}

# Reference the validated location in a region resource.
# This ensures the location exists before creating the region.
data "gamefabric_environment" "prod" {
  display_name = "Production"
}

resource "gamefabric_region" "us" {
  name         = "us-east"
  display_name = "US East"
  environment  = data.gamefabric_environment.prod.name

  types = {
    cloud = {
      # Reference the validated location
      locations  = [data.gamefabric_location.us_east.name]
      scheduling = "Distributed"
    }
  }
}
