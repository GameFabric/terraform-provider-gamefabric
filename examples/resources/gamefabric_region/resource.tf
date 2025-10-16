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

  # Annotations for storing additional metadata (optional)
  annotations = {}

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
          value = "Europe"           # The environment variable value
        },
        {
          name = "POD_NAME"
          value_from = {
            field_ref = {
              field_path = "metadata.name"
            }
          }
        }
        {
            name = "CONFIG_FILE"
            value_from = {
              config_file_key_ref = {
                name = "eu-config" # Reference a ConfigMap named "eu-config"
              }
            }
        }
      ]

      # Scheduling strategy for distributing workloads (optional, defaults to "Packed")
      # - "Packed": Fills up locations before moving to the next (cost-optimized)
      # - "Distributed": Spreads workloads evenly across locations
      scheduling = "Packed"
    }
  }
}
