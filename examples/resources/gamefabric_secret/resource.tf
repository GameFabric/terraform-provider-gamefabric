resource "gamefabric_secret" "db_credentials" {
  name        = "examplesecret"
  environment = data.gamefabric_environment.dev.name
  description = "Backend database credentials"
  labels = {
    component = "database"
  }
  data = {
    db_user     = "dbuser123"
    db_password = "abcdef123456!#$&'*+-=?^_`{|}~"
  }
}
