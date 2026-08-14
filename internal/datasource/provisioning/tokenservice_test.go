package provisioning_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTokenServiceDataSource(t *testing.T) {
	ts := &provisioningv1beta1.TokenService{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-token-service",
			Labels: map[string]string{"team": "platforms"},
		},
		Spec: provisioningv1beta1.TokenServiceSpec{
			Environment: provisioningv1beta1.TokenServiceEnvProd,
			Game: provisioningv1beta1.TokenServiceGameSpec{
				Name: "my-game",
				Platforms: map[string]provisioningv1beta1.TokenServicePlatformSpec{
					"pc": {
						GameClientTokenKeys: []provisioningv1beta1.TokenServiceKeySpec{
							{Key: "MDEyMzQ1Njc4OWFiY2RlZg==", SigningAlgorithm: provisioningv1beta1.TokenServiceSigningAlgorithmHS256},
						},
					},
				},
			},
		},
		Status: provisioningv1beta1.TokenServiceStatus{
			State:        provisioningv1beta1.TokenServiceStateAvailable,
			Hostname:     "my-game.tokens.example.com",
			PlatformKeys: `{"pc":["key1","key2"]}`,
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, ts)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_steelshield_tokenservice" "test" {
  name = "test-token-service"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "name", "test-token-service"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "development_mode", "false"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "game_name", "my-game"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "labels.team", "platforms"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "state", "Available"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "hostname", "my-game.tokens.example.com"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "platform_key", "key1"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "platforms.pc.game_client_token_keys.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "platforms.pc.game_client_token_keys.0.signing_algorithm", "HS256"),
				),
			},
		},
	})
}

// TestTokenServiceDataSource_EOS verifies the "eos" platform entry (STS-2888 bridge) is
// hidden from the platforms attribute.
func TestTokenServiceDataSource_EOS(t *testing.T) {
	ts := &provisioningv1beta1.TokenService{
		ObjectMeta: metav1.ObjectMeta{Name: "eos-token-service"},
		Spec: provisioningv1beta1.TokenServiceSpec{
			Environment: provisioningv1beta1.TokenServiceEnvProd,
			Game: provisioningv1beta1.TokenServiceGameSpec{
				Name: "my-game",
				Platforms: map[string]provisioningv1beta1.TokenServicePlatformSpec{
					"eos": {},
				},
			},
			EOS: []provisioningv1beta1.TokenServiceEOSSpec{{ClientID: "client-id"}},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, ts)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_steelshield_tokenservice" "test" {
  name = "eos-token-service"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "eos.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservice.test", "platforms.%", "0"),
				),
			},
		},
	})
}

func TestTokenServiceDataSource_NotFound(t *testing.T) {
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_steelshield_tokenservice" "test" {
  name = "does-not-exist"
}
`,
				ExpectError: regexp.MustCompile(`Token Service Not Found`),
			},
		},
	})
}
