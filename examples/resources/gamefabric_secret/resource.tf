# Example 1: Secret with persistent data (will be stored in state)
resource "gamefabric_secret" "db_credentials" {
  name        = "examplesecret"
  environment = data.gamefabric_environment.dev.name
  description = "Backend database credentials"
  labels = {
    game       = "my_first_game"
    component  = "database"
  }
  data = {
    db_user     = "dbuser123"
    db_password = "abcdef123456!#$&'*+-=?^_`{|}~"
  }
}

# Example 2: Secret with write-only data (will NOT be stored in state)
resource "gamefabric_secret" "api_key" {
  name        = "examplesecret"
  environment = data.gamefabric_environment.prod.name
  description = "API key for external service"
  labels = {
    game       = "my_second_game"
    component  = "api"
  }

  # Write-only attributes require Terraform v1.11 and later.
  data_wo = {
    api_key    = "sk-1234567890abcdef"
    api_secret = "secret-xyz-9876543210"
  }
}
