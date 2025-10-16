# Configure Terraform and required providers
terraform {
  required_providers {
    gamefabric = {
      source  = "gamefabric/gamefabric" # The GameFabric Terraform provider from the Terraform registry
      version = ">=1.0.0"                # Minimum required version
    }
  }
  required_version = ">= 1.0.0" # Minimum Terraform version required
}

# Configure the GameFabric provider with authentication credentials
provider "gamefabric" {
  # The GameFabric API host URL (optional)
  # If not provided, the host will be derived from customer_id as: {customer_id}.gamefabric.dev
  # Can be overridden via environment variable: GAMEFABRIC_HOST
  host = "<your gamefabric installation host>"

  # The customer ID for your GameFabric installation (required)
  # This is used to identify your organization/tenant
  # If host is not specified, it will default to: {customer_id}.gamefabric.dev
  # Can be set via environment variable: GAMEFABRIC_CUSTOMER_ID
  customer_id = "<your customer id>"

  # The service account username for authentication (required)
  # This is the username of a service account with appropriate permissions
  # Can be set via environment variable: GAMEFABRIC_SERVICE_ACCOUNT
  service_account = "<your username>"

  # The password for the service account (required, sensitive)
  # This value is marked as sensitive and won't appear in logs or output
  # Can be set via environment variable: GAMEFABRIC_PASSWORD
  password = "<your password>"
}

# Example configuration using environment variables instead of hardcoded values:
# provider "gamefabric" {
#   # All values can be provided via environment variables:
#   # export GAMEFABRIC_CUSTOMER_ID="your-customer-id"
#   # export GAMEFABRIC_SERVICE_ACCOUNT="your-service-account"
#   # export GAMEFABRIC_PASSWORD="your-password"
#   # export GAMEFABRIC_HOST="your-host"  # Optional
# }
