resource "gamefabric_service_account" "example" {
  name = "my-service-account"
}

resource "gamefabric_service_account_password" "example" {
  service_account = gamefabric_service_account.example.name
}

output "password"  {
  value = nonsensitive(gamefabric_service_account_password.example.password)
}