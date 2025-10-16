# Create a development environment with basic configuration.
# An environment is a logical isolation boundary for resources like regions, armadas, etc.
# Each environment has its own set of resources and can have different configurations.
resource "gamefabric_environment" "dev" {
  # The unique name of the environment (required)
  # Must be unique across all environments and is limited to 4 characters
  # This is the technical identifier used in API calls and resource references
  name = "dev"

  # The user-friendly display name (required)
  # Can contain spaces and special characters, used in UI and reports
  display_name = "Development"

  # Optional description explaining the purpose of the environment
  description = "Development environment for testing and experimentation"

  # Labels for organizing and categorizing environments (optional)
  # Labels are key-value pairs that can be used for filtering with the gamefabric_environments data source
  # They help organize environments and enable policy-based operations
  labels = {
    environment_type = "non-prod"
    stage            = "dev"
    team             = "platform"
    cost_center      = "engineering"
  }
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
}

# Create a staging environment
resource "gamefabric_environment" "stg" {
  name         = "stg"
  display_name = "Staging"
  description  = "Staging environment for pre-production testing"

  labels = {
    environment_type = "non-prod"
    stage            = "staging"
  }
}

# Create a QA environment
resource "gamefabric_environment" "qa" {
  name         = "qa"
  display_name = "Quality Assurance"
  description  = "QA environment for automated and manual testing"

  labels = {
    environment_type = "non-prod"
    stage            = "qa"
    testing          = "automated"
  }
}

# Output environment information for use in other resources
output "dev_environment_name" {
  description = "The name of the development environment"
  value       = gamefabric_environment.dev.name
}

output "all_environment_labels" {
  description = "Labels for all created environments"
  value = {
    dev  = gamefabric_environment.dev.labels
    prod = gamefabric_environment.prod.labels
    stg  = gamefabric_environment.stg.labels
    qa   = gamefabric_environment.qa.labels
  }
}
