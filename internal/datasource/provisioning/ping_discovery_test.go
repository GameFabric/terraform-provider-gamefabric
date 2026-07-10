package provisioning_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPingDiscovery(t *testing.T) {
	t.Parallel()

	pd := &provisioningv1beta1.PingDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pd",
		},
		Status: provisioningv1beta1.PingDiscoveryStatus{
			URL:    "https://ping.example.com",
			Tokens: []string{"pd-token-old", "pd-token-new"},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, pd)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_ping_discovery" "test" {
  name = "test-pd"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "name", "test-pd"),
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "url", "https://ping.example.com"),
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "active_token", "pd-token-new"),
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "tokens.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "tokens.0", "pd-token-old"),
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "tokens.1", "pd-token-new"),
				),
			},
		},
	})
}

func TestPingDiscovery_NoTokens(t *testing.T) {
	t.Parallel()

	pd := &provisioningv1beta1.PingDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name: "empty-pd",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, pd)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_ping_discovery" "test" {
  name = "empty-pd"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "name", "empty-pd"),
					resource.TestCheckNoResourceAttr("data.gamefabric_ping_discovery.test", "url"),
					resource.TestCheckNoResourceAttr("data.gamefabric_ping_discovery.test", "active_token"),
					resource.TestCheckResourceAttr("data.gamefabric_ping_discovery.test", "tokens.#", "0"),
				),
			},
		},
	})
}
