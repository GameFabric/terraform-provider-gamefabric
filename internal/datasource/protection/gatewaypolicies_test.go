package protection_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGatewayPolicies(t *testing.T) {
	t.Parallel()

	pol := &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, pol)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_gatewaypolicies" "test1" {
  label_filter = {
	env = "prod"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicies.test1", "label_filter.env", "prod"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicies.test1", "names.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicies.test1", "names.0", "test-policy"),
				),
			},
		},
	})
}

func TestEnvironments_AllowsGettingAll(t *testing.T) {
	t.Parallel()

	pol := &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, pol)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_gatewaypolicies" "test1" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicies.test1", "names.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicies.test1", "names.0", "test-policy"),
				),
			},
		},
	})
}
