# Get an environment by its display name.
# The display name is the user-friendly name shown in the UI.
# Note: Either 'name' or 'display_name' must be specified, but not both.
data "gamefabric_environment" "dev" {
  display_name = "Development"
}

# Alternatively, get an environment by its unique name.
# The name is the technical identifier (limited to 4 characters).
data "gamefabric_environment" "prod_by_name" {
  name = "prod"
}

# Access environment attributes after retrieval.
# The environment data source returns name, display_name, labels, and description.
output "dev_environment_info" {
  description = "Information about the development environment"
  value = {
    name         = data.gamefabric_environment.dev.name
    display_name = data.gamefabric_environment.dev.display_name
    labels       = data.gamefabric_environment.dev.labels
    description  = data.gamefabric_environment.dev.description
  }
}

# Use the environment data source in other resources.
# This is useful when referencing environments that already exist.
data "gamefabric_environment" "production" {
  name = "prod"
}

resource "gamefabric_region" "us_east" {
  name         = "us-east"
  display_name = "US East"
  # Reference the environment from the data source
  environment = data.gamefabric_environment.production.name

  types = {
    cloud = {
      locations  = ["us-east-1"]
      scheduling = "Packed"
    }
  }
}
