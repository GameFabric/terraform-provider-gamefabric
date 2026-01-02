# gamefabric_service_account_password Resource

Manages a service account password resource in GameFabric.

## Example Usage

```terraform
resource "gamefabric_service_account" "example" {
  name = "my-service-account"
}

resource "gamefabric_service_account_password" "example" {
  service_account = gamefabric_service_account.example.name
  labels = {
    environment = "production"
    team        = "platform"
  }
}
```

## Argument Reference

- `service_account` - (Required) The name of the service account. Forces replacement when changed.
- `labels` - (Optional) A map of labels to assign to the service account password. Forces replacement when changed.

## Attribute Reference

- `id` - The ID of the service account password resource (same as `service_account`).
- `password_wo` - (Computed, Sensitive) The password for the service account. This is a write-only attribute that is only available immediately after creation and is not retrievable from the API afterwards.

## Import

Service account passwords can be imported using the service account name:

```shell
terraform import gamefabric_service_account_password.example my-service-account
```

