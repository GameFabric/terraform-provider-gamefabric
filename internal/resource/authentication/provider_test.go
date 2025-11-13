package authentication_test

import (
	"fmt"
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	name := "test-provider"
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: testResourceProviderConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "display_name", "My Provider"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.issuer", "https://example.com"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.client_id", "my-client-id"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.client_secret", "my-secr3t"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.redirect_uri", "https://example.com/redirect"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.#", "3"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.0", "openid"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.1", "email"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.2", "profile"),
				),
			},
			{
				ResourceName:      "gamefabric_authentication_provider.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceProviderConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "display_name", "My Provider"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.%", "21"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.issuer", "https://example.com"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.client_id", "my-client-id"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.client_secret", "my-secr3t"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.redirect_uri", "https://example.com/redirect"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.acr_values.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.acr_values.0", "urn:mace:incommon:iap:silver"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.basic_auth_unsupported", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_mapping.email", "email"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_mapping.groups", "groups"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_mapping.preferred_username", "preferred_username"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.%", "2"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.filter_group_claims.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.filter_group_claims.groups_filter", "^dev-.*$"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.0.claims.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.0.claims.0", "claim1"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.0.claims.1", "claim2"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.0.delimiter", "-"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.0.clear_delimiter", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.claim_modification.new_group_from_claims.0.prefix", "role_"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.get_user_info", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.hosted_domains.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.hosted_domains.0", "example.org"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.hosted_domains.1", "example.com"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.insecure_enable_groups", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.insecure_skip_email_verified", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.insecure_skip_verify", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.override_claim_mapping", "true"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.prompt_type", "consent"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.provider_discovery_override.%", "3"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.provider_discovery_override.auth_url", "https://example.org/auth"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.provider_discovery_override.jwks_url", "https://example.org/jwks"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.provider_discovery_override.token_url", "https://example.org/token"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.root_ca_certificates.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.root_ca_certificates.0", "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.#", "7"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.0", "openid"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.1", "email"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.2", "profile"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.3", "groups"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.4", "federated:id"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.5", "offline_access"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.scopes.6", "audience:server:client_id:(client-id)"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.user_id_key", "email"),
					resource.TestCheckResourceAttr("gamefabric_authentication_provider.test", "oidc.user_name_key", "user_name_key"),
				),
			},
		},
	})
}

func testResourceProviderConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_authentication_provider" "test" {
  name = %q
  display_name = "My Provider"
  oidc = {
    issuer = "https://example.com"
    client_id = "my-client-id"
    client_secret = "my-secr3t"
    redirect_uri = "https://example.com/redirect"
    scopes = ["openid", "email", "profile"]
  }
}`, name)
}

func testResourceProviderConfigFull(name string) string {
	return fmt.Sprintf(`resource "gamefabric_authentication_provider" "test" {
  name = %q
  display_name = "My Provider"
  oidc = {
    issuer = "https://example.com"
    client_id = "my-client-id"
    client_secret = "my-secr3t"
    redirect_uri = "https://example.com/redirect"
    acr_values = ["urn:mace:incommon:iap:silver"]
    basic_auth_unsupported = true
    claim_mapping = {
      email = "email"
      groups = "groups"
      preferred_username = "preferred_username"
    }
    claim_modification = {
      filter_group_claims = {
       groups_filter = "^dev-.*$"
      }
      new_group_from_claims = [{
       claims = ["claim1", "claim2"]
       delimiter = "-"
       clear_delimiter = true
       prefix = "role_"
      }]
    }
    get_user_info = true
    hosted_domains = ["example.org", "example.com"]
    insecure_enable_groups = true
    insecure_skip_email_verified = true
    insecure_skip_verify = true
    override_claim_mapping = true
    prompt_type = "consent"
    provider_discovery_override = {
      auth_url = "https://example.org/auth"
      jwks_url = "https://example.org/jwks"
      token_url = "https://example.org/token"
    }
    root_ca_certificates = ["-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"]
    scopes = ["openid", "email", "profile", "groups", "federated:id", "offline_access", "audience:server:client_id:(client-id)"]
    user_id_key = "email"
    user_name_key = "user_name_key"
  }
}`, name)
}
