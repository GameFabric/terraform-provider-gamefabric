resource "gamefabric_environment" "dev" {
  name = "dev" # The name must be unique and is limited to 4 characters.
  labels = {
    # Labels are optional, but can be used to organize environments and be used as filters for in the gamefabric_environments data source.
    environment_type = "non-prod"
    stage            = "dev"
  }
  description  = "All development resources"
  display_name = "Development"
}
