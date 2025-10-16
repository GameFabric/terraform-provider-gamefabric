# Lookup the locations for the baremetal type in Europe.
# This data source will return all locations that match the specified criteria.
# Locations represent physical or cloud infrastructure sites where resources can be deployed.
data "gamefabric_locations" "eu_baremetal" {
  type      = "baremetal" # Filter by infrastructure type (e.g., "baremetal", "cloud", "edge")
  continent = "europe"    # Filter by continent
  #city      = "frankfurt"     # Optional: Filter by specific city
  #name_regex = ".*gcp"         # Optional: Filter by location name pattern
}

# Get all environments to setup the region consistently across environments.
# This pattern is useful when you want to create identical regions in all environments.
data "gamefabric_environments" "all" {
}

# Create a region for Europe that uses the baremetal type.
# A region represents a logical grouping of locations where game servers or services can be deployed.
resource "gamefabric_region" "europe" {
  # Use for_each to create the region in all environments
  for_each = { for env in data.gamefabric_environments.all.environments : env.name => env }

  # The unique name of the region within the environment (required)
  # Must be unique within the environment
  name = "europe"

  # The user-friendly display name for the region (required)
  display_name = "Europe"

  # The environment this region belongs to (required)
  # Regions are scoped to a specific environment
  environment = each.value.name

  # Labels for organizing and categorizing regions (optional)
  # Labels can be used for filtering and organization
  labels = {
    continent = "europe"
    region_type = "production"
  }

  # Optional description for the region
  description = "European region for game server deployments"

  # Define the types of infrastructure available in this region (required)
  # At least one type must be specified
  types = {
    # The key (e.g., "baremetal") defines the infrastructure type name
    baremetal = {
      # List of location names where this type is available (required)
      # Locations must exist and be accessible
      locations = data.gamefabric_locations.eu_baremetal.names

      # Optional environment variables to set on all containers in this region/type
      # These are injected into game server containers
      env = [
        {
          name  = "GAME_REGION_NAME" # The environment variable name
          value = "Europe"            # The environment variable value
        },
        {
          name  = "REGION_CODE"
          value = "EU"
        }
      ]

      # Scheduling strategy for distributing workloads (optional, defaults to "Packed")
      # - "Packed": Fills up locations before moving to the next (cost-optimized)
      # - "Distributed": Spreads workloads evenly across locations (redundancy-optimized)
      scheduling = "Packed"
    }
  }
}

# Example with multiple infrastructure types in the same region
resource "gamefabric_region" "us_multi" {
  name         = "us-multi"
  display_name = "US Multi-Type"
  environment  = data.gamefabric_environments.all.environments[0].name

  # A region can support multiple infrastructure types
  types = {
    # Baremetal for high-performance game servers
    baremetal = {
      locations  = ["us-east-1-nyc", "us-west-1-lax"]
      scheduling = "Distributed"
    }

    # Cloud for scalable auxiliary services
    cloud = {
      locations  = ["us-east-1-aws", "us-west-1-gcp"]
      scheduling = "Packed"
      env = [
        {
          name  = "DEPLOYMENT_TYPE"
          value = "cloud"
        }
      ]
    }
  }
}
