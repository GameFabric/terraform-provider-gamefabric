package provisioning_test

import (
	"fmt"
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMZ4hj8vG6Vh7A7v9XlPQ2r3s5N2R8sX
1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWCAwEA
AQ==
-----END PUBLIC KEY-----`

func TestTokenService_CustomKeys(t *testing.T) {
	name := "custom-keys-service"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceTokenServiceConfigCustomKeys(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "development_mode", "false"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "game_name", "my-game"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "labels.team", "platforms"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "platforms.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "platforms.pc.game_client_token_keys.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "platforms.pc.game_client_token_keys.0.signing_algorithm", "HS256"),
				),
			},
			{
				ResourceName:      "gamefabric_steelshield_tokenservice.test",
				ImportState:       true,
				ImportStateVerify: true,
				// state is set by backend reconciliation, which the fake API does not simulate.
				// development_mode: on import, Read normalizes the model against the (null)
				// import state, and normalize clobbers a false bool to null, so the imported
				// value can't be verified.
				ImportStateVerifyIgnore: []string{"state", "development_mode"},
			},
			{
				Config: testResourceTokenServiceConfigCustomKeysUpdatedLabels(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "labels.team", "platforms-updated"),
				),
			},
		},
	})
}

func TestTokenService_JWKS(t *testing.T) {
	name := "jwks-service"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceTokenServiceConfigJWKS(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "jwks.url", "https://tokens.mygame.com/.well-known/jwks.json"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "platforms.%", "1"),
				),
			},
		},
	})
}

func TestTokenService_EOS(t *testing.T) {
	name := "eos-service"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceTokenServiceConfigEOS(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "eos.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "eos.0.client_id", "1234567890abcdef1234567890abcdef"),
					// token_types defaults to ["connect"] when omitted.
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "eos.0.token_types.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "eos.0.token_types.0", "connect"),
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "platforms.%", "0"),
					// STS-2888 bridge: API object still gets an "eos" platform entry.
					testResourceTokenServiceCheckAPIPlatforms(t, cs, name, "eos"),
				),
			},
		},
	})
}

func TestTokenService_DefaultDevelopmentMode(t *testing.T) {
	name := "default-mode-service"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceTokenServiceConfigDefaultMode(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "name", name),
					// development_mode is omitted from the config, so it must default to false.
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "development_mode", "false"),
				),
			},
		},
	})
}

func TestTokenService_UpdateDevelopmentMode(t *testing.T) {
	name := "update-mode-service"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceTokenServiceConfigCustomKeys(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "development_mode", "false"),
				),
			},
			{
				// Flipping development_mode is an in-place update (PATCH), not a replacement.
				Config: withDevelopmentMode(testResourceTokenServiceConfigCustomKeys(name), "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_steelshield_tokenservice.test", "development_mode", "true"),
				),
			},
		},
	})
}

func TestTokenService_ValidatesGameName(t *testing.T) {
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      withGameName(testResourceTokenServiceConfigCustomKeys("bad-game-name"), "Not_Valid"),
				ExpectError: regexp.MustCompile(`game_name must contain only lowercase letters, numbers, and hyphens`),
			},
		},
	})
}

func TestTokenService_ValidatesModeExclusivity(t *testing.T) {
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      testResourceTokenServiceConfigConflictingModes("conflict-service"),
				ExpectError: regexp.MustCompile(`Conflicting Auth Mode Configuration`),
			},
		},
	})
}

func TestTokenService_ValidatesPlatformName(t *testing.T) {
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      withPlatform(testResourceTokenServiceConfigCustomKeys("bad-platform"), "nintendo64"),
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*nintendo64`),
			},
		},
	})
}

func TestTokenService_ValidatesKeyMaterial(t *testing.T) {
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceTokenServiceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      withKey(testResourceTokenServiceConfigCustomKeys("bad-key-material"), "not-base64!!!", "HS256"),
				ExpectError: regexp.MustCompile(`not valid base64`),
			},
			{
				Config:      withKey(testResourceTokenServiceConfigCustomKeys("empty-key-material"), "", "HS256"),
				ExpectError: regexp.MustCompile(`(?s)key.*string length must be\s+at least 1`),
			},
		},
	})
}

type tokenServiceConfig = string

func withGameName(c tokenServiceConfig, gameName string) tokenServiceConfig {
	return regexp.MustCompile(`game_name\s*=\s*"[^"]*"`).ReplaceAllString(c, fmt.Sprintf(`game_name = %q`, gameName))
}

func withPlatform(c tokenServiceConfig, platform string) tokenServiceConfig {
	return regexp.MustCompile(`pc\s*=\s*\{`).ReplaceAllString(c, platform+" = {")
}

func withKey(c tokenServiceConfig, key, alg string) tokenServiceConfig {
	s := regexp.MustCompile(`key\s*=\s*"[^"]*"`).ReplaceAllString(c, fmt.Sprintf(`key = %q`, key))
	s = regexp.MustCompile(`signing_algorithm\s*=\s*"[^"]*"`).ReplaceAllString(s, fmt.Sprintf(`signing_algorithm = %q`, alg))
	return s
}

func withDevelopmentMode(c tokenServiceConfig, mode string) tokenServiceConfig {
	return regexp.MustCompile(`development_mode\s*=\s*\w+`).ReplaceAllString(c, fmt.Sprintf(`development_mode = %s`, mode))
}

func testResourceTokenServiceConfigCustomKeys(name string) tokenServiceConfig {
	return tokenServiceConfig(fmt.Sprintf(`resource "gamefabric_steelshield_tokenservice" "test" {
  name             = %q
  development_mode = false
  game_name        = "my-game"

  labels = {
    team = "platforms"
  }

  platforms = {
    pc = {
      game_client_token_keys = [
        {
          key               = "MDEyMzQ1Njc4OWFiY2RlZg=="
          signing_algorithm = "HS256"
        }
      ]
    }
  }
}`, name))
}

func testResourceTokenServiceConfigCustomKeysUpdatedLabels(name string) tokenServiceConfig {
	return tokenServiceConfig(fmt.Sprintf(`resource "gamefabric_steelshield_tokenservice" "test" {
  name             = %q
  development_mode = false
  game_name        = "my-game"

  labels = {
    team = "platforms-updated"
  }

  platforms = {
    pc = {
      game_client_token_keys = [
        {
          key               = "MDEyMzQ1Njc4OWFiY2RlZg=="
          signing_algorithm = "HS256"
        }
      ]
    }
  }
}`, name))
}

func testResourceTokenServiceConfigJWKS(name string) tokenServiceConfig {
	return tokenServiceConfig(fmt.Sprintf(`resource "gamefabric_steelshield_tokenservice" "test" {
  name             = %q
  development_mode = false
  game_name        = "my-game"

  platforms = {
    pc = {}
  }

  jwks = {
    url = "https://tokens.mygame.com/.well-known/jwks.json"
  }
}`, name))
}

func testResourceTokenServiceConfigEOS(name string) tokenServiceConfig {
	return tokenServiceConfig(fmt.Sprintf(`resource "gamefabric_steelshield_tokenservice" "test" {
  name             = %q
  development_mode = false
  game_name        = "my-game"

  eos = [
    {
      client_id     = "1234567890abcdef1234567890abcdef"
      product_id    = "abc123"
      deployment_id = "def456"
      sandbox_id    = "ghi789"
    },
  ]
}`, name))
}

func testResourceTokenServiceConfigDefaultMode(name string) tokenServiceConfig {
	return tokenServiceConfig(fmt.Sprintf(`resource "gamefabric_steelshield_tokenservice" "test" {
  name      = %q
  game_name = "my-game"

  platforms = {
    pc = {
      game_client_token_keys = [
        {
          key               = "MDEyMzQ1Njc4OWFiY2RlZg=="
          signing_algorithm = "HS256"
        }
      ]
    }
  }
}`, name))
}

func testResourceTokenServiceConfigConflictingModes(name string) tokenServiceConfig {
	return tokenServiceConfig(fmt.Sprintf(`resource "gamefabric_steelshield_tokenservice" "test" {
  name             = %q
  development_mode = false
  game_name        = "my-game"

  platforms = {
    pc = {
      game_client_token_keys = [
        {
          key               = "MDEyMzQ1Njc4OWFiY2RlZg=="
          signing_algorithm = "HS256"
        }
      ]
    }
  }

  jwks = {
    url = "https://tokens.mygame.com/.well-known/jwks.json"
  }
}`, name))
}

func testResourceTokenServiceDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_steelshield_tokenservice" {
				continue
			}

			resp, err := cs.ProvisioningV1Beta1().TokenServices().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil && resp.Name == rs.Primary.ID {
				return fmt.Errorf("token service still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}

// testResourceTokenServiceCheckAPIPlatforms asserts the API object's platform keys directly,
// bypassing Terraform state.
func testResourceTokenServiceCheckAPIPlatforms(t *testing.T, cs clientset.Interface, name string, wantKeys ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		obj, err := cs.ProvisioningV1Beta1().TokenServices().Get(t.Context(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("could not get token service: %w", err)
		}
		if len(obj.Spec.Game.Platforms) != len(wantKeys) {
			return fmt.Errorf("expected platforms %v, got %v", wantKeys, obj.Spec.Game.Platforms)
		}
		for _, k := range wantKeys {
			if _, ok := obj.Spec.Game.Platforms[k]; !ok {
				return fmt.Errorf("expected platform %q, got %v", k, obj.Spec.Game.Platforms)
			}
		}
		return nil
	}
}
