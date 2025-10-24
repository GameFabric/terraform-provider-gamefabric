resource "gamefabric_configfile" "content" {
  environment = data.gamefabric_environment.prod.name
  name        = "exampleconfig"
  data        = <<EOT
[/Script/Engine.GameSession]
MaxPlayers=64

[/Script/OnlineSubsystemUtils.IpNetDriver]
NetServerMaxTickRate=60
MaxInternetClientRate=25000
EOT
}

resource "gamefabric_configfile" "content_from_file" {
  environment = data.gamefabric_environment.prod.name
  name        = "exampleconfig"
  data        = file("${path.module}/hello.txt")
}
