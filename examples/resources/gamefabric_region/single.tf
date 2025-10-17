data "gamefabric_environment" "prod" {
  name = "prod"
}

data "gamefabric_locations" "cloud_europe" {
  type      = "cloud"
  continent = "europe"
}

data "gamefabric_locations" "baremetal_europe" {
  type      = "baremetal"
  continent = "europe"
}

resource "gamefabric_region" "europe" {
  name         = "europe"
  display_name = "Europe"
  environment  = data.gamefabric_environment.prod.name
  types = {
    "baremetal" = {
      locations = data.gamefabric_locations.baremetal_europe.names
      envs = [
        {
          name = "REGION"
          value_from = {
            field_ref = {
              field_path = "metadata.regionName" # e.g., "europe"
            }
          }
        },
        {
          name = "REGION_TYPE"
          value_from = {
            field_ref = {
              field_path = "metadata.regionTypeName" # e.g., "baremetal"
            }
          }
        },
        {
          name  = "BACKEND_URL"
          value = "https://mybackend.eu.example.com"
        },
        {
          name = "BACKEND_TOKEN"
          value_from = {
            config_file_key_ref = {
              name = "eu-token" # Reference a ConfigFile named "eu-token"
            }
          }
        }
      ]
      scheduling = "Distributed"
    },
    "cloud" = {
      locations = data.gamefabric_locations.cloud_europe.names
      envs = [
        {
          name = "REGION"
          value_from = {
            field_ref = {
              field_path = "metadata.regionName" # e.g., "europe"
            }
          }
        },
        {
          name = "REGION_TYPE"
          value_from = {
            field_ref = {
              field_path = "metadata.regionTypeName" # e.g., "cloud"
            }
          }
        },
        {
          name  = "BACKEND_URL"
          value = "https://mybackend.eu.example.com"
        },
        {
          name = "BACKEND_TOKEN"
          value_from = {
            config_file_key_ref = {
              name = "eu-token" # Reference a ConfigFile named "eu-token"
            }
          }
        }
      ]
      scheduling = "Packed"
    }
  }
}
