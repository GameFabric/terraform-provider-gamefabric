# Get all environments without any filtering.
# This returns a list of all environments in the system.
data "gamefabric_environments" "all" {
}

# Get environments filtered by labels.
# Labels are key-value pairs used to organize and categorize environments.
# Only environments that have ALL specified labels with matching values are returned.
data "gamefabric_environments" "non_prod" {
  label_filter = {
    environment_type = "non-prod"
  }
}

# Get environments with multiple label filters.
# This example filters for development environments in a specific stage.
data "gamefabric_environments" "dev_stage" {
  label_filter = {
    environment_type = "non-prod"
    stage            = "dev"
  }
}

# Access the list of environments returned by the data source.
# Each environment includes name, display_name, labels, and description.
output "all_environment_names" {
  description = "List of all environment names"
  value       = [for env in data.gamefabric_environments.all.environments : env.name]
}

# Use the environments data source in a for_each to create resources.
# This is useful when you need to create identical resources across multiple environments.
resource "gamefabric_region" "example" {
  for_each = { for env in data.gamefabric_environments.non_prod.environments : env.name => env }

  name         = "example-region"
  display_name = "Example Region"
  environment  = each.value.name

  types = {
    baremetal = {
      locations  = ["location-1"]
      scheduling = "Packed"
    }
  }
}
