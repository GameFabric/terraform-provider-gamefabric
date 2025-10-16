resource "gamefabric_environment" "prod" {
  name         = "prod"
  display_name = "Production"
  description  = "Production environment for live game servers"

  labels = {
    environment_type = "production"
    criticality      = "high"
  }
}
