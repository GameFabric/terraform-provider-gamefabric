data "gamefabric_environment" "dev" {
  display_name = "Development"
}

# Get the region by display name.
data "gamefabric_region" "name" {
  display_name = "Europe"
  environment  = data.gamefabric_environment.dev.name
}
