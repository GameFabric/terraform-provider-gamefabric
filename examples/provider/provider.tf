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
  host            = "<your gamefabric installation host>" # env var: GAMEFABRIC_HOST (optional, overrides the customer based host)
  customer_id     = "<your customer id>"                  # env var: GAMEFABRIC_CUSTOMER_ID (optional if set host is optional and we default to $customer_id.gamefabric.dev)
  service_account = "<your username>"                     # env var: GAMEFABRIC_SERVICE_ACCOUNT
  password        = "<your password>"                     # env var: GAMEFABRIC_PASSWORD
}
