data "gamefabric_configfile" "base_configuration" {
  environment = gamefabric_environment.prod.name
  name        = "baseconfiguration"
}
