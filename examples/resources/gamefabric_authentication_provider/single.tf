resource "gamefabric_authentication_provider" "company_oidc" {
  name         = "company-oidc-provider"
  display_name = "Company OIDC Authentication"

  oidc = {
    issuer        = "https://auth.company.example.com"
    client_id     = "gamefabric-client-id"
    client_secret = "your-client-secret-here"
    redirect_uri  = "https://gamefabric.company.example.com/callback"

    scopes = [
      "openid",
      "profile",
      "email",
      "groups"
    ]

    user_id_key   = "sub"
    user_name_key = "preferred_username"

    insecure_enable_groups       = true
    insecure_skip_email_verified = false
    get_user_info                = true

    allowed_groups = [
      "gamefabric-admins",
      "gamefabric-developers"
    ]

    claim_mapping = {
      email              = "email"
      groups             = "groups"
      preferred_username = "preferred_username"
    }

    claim_modification = {
      filter_group_claims = {
        groups_filter = "^gamefabric-.*"
      }
    }
  }
}
