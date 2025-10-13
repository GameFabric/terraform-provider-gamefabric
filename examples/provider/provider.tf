terraform {
  required_providers {
    gamefabric = {
      source  = "gamefabric/gamefabric"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 1.0.0"
}

provider "gamefabric" {
  host            = "<your enterprise console host url>" # $customer_id.gamefabric.dev
  customer_id     = "<your customer id>"                 # $customer_id.gamefabric.dev
  service_account = "<your username>"
  password        = "<your password>" # we recommend to use the GAMEFABRIC_PASSWORD environment variable
}
