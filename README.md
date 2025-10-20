# Terraform Provider for GameFabric

The GameFabric provider is a Terraform plugin for managing GameFabric resources.

## Installation

This provider is distributed through the Terraform Registry.

## Building

Building from source requires access to private GameFabric repositories and is only available to GameFabric developers.

## Using the Provider

```terraform
terraform {
  required_providers {
    gamefabric = {
      source = "gamefabric/gamefabric"
    }
  }
}

provider "gamefabric" {
  customer_id     = "<your customer id>"
  service_account = "ExampleServiceAccount@ec.nitrado.systems"
}
```