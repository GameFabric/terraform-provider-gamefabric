# Example 3: Secret with write-only ephemeral data (will NOT be stored in state)
#
# Ephemeral resources are available in Terraform v1.10 and later.
ephemeral "random_password" "db_password" {
  length           = 16
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "gamefabric_secret" "db_credentials" {
  name        = "db-creds"
  environment = data.gamefabric_environment.prod.name
  description = "Database credentials"

  # Write-only attributes require Terraform v1.11 and later.
  data_wo = {
    db_user    = "secr3t_user"
    api_secret = ephemeral.random_password.db_password.result
  }
  data_wo_version = 1
}
