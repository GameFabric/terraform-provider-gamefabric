# Create a development environment with basic configuration.
# An environment is a logical isolation boundary for resources like regions, armadas, etc.
# Each environment has its own set of resources and can have different configurations.
resource "gamefabric_environment" "dev" {
  name         = "dev"         # Required: Unique identifier.
  display_name = "Development" # Required: User-friendly display name for the environment.
}

# Create a production environment with minimal labels
resource "gamefabric_environment" "prod" {
  name         = "prod" # 4-character limit
  display_name = "Production"
  description  = "Production environment for live game servers"

  labels = {
    environment_type = "production"
    criticality      = "high"
  }

  annotations = {
    owner = "ops-team"
  }
}
