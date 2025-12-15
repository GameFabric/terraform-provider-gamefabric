data "gamefabric_secret" "base_secret" {
  environment = gamefabric_environment.prod.name
  name        = "basesecret"
}
