# First, get the environment that contains the regions.
# Regions are always scoped to a specific environment.
data "gamefabric_environment" "dev" {
  display_name = "Development"
}

# Get all regions in the development environment without filtering.
# This returns all regions, regardless of their labels or other attributes.
data "gamefabric_regions" "all" {
  environment = data.gamefabric_environment.dev.name
}

# Get regions filtered by labels.
# Labels are key-value pairs used to organize and categorize regions.
# Only regions that have ALL specified labels with matching values are returned.
data "gamefabric_regions" "us" {
  environment = data.gamefabric_environment.dev.name
  label_filter = {
    continent = "us"
  }
}

# Get regions with multiple label filters.
# This example filters for production regions in Europe.
data "gamefabric_regions" "eu_prod" {
  environment = data.gamefabric_environment.dev.name
  label_filter = {
    continent   = "europe"
    region_type = "production"
  }
}

# Access the list of regions returned by the data source.
# The regions data source returns a list of region objects with full details.
output "all_region_names" {
  description = "Names of all regions in the development environment"
  value       = [for region in data.gamefabric_regions.all.regions : region.name]
}

output "us_region_details" {
  description = "Detailed information about US regions"
  value = [
    for region in data.gamefabric_regions.us.regions : {
      name         = region.name
      display_name = region.display_name
      types        = keys(region.types)
    }
  ]
}
