package protection_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGatewayPolicy(t *testing.T) {
	pol := &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test-gateway-policy"},
		Spec: protectionv1.GatewayPolicySpec{
			DisplayName:      "My Gateway Policy",
			Description:      "My Gateway Policy Description",
			DestinationCIDRs: []string{"0.0.0.0/0"},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, pol)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_gatewaypolicy" "test1" {
  name = "test-gateway-policy"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test1", "name", "test-gateway-policy"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test1", "display_name", "My Gateway Policy"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test1", "description", "My Gateway Policy Description"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test1", "destination_cidrs.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test1", "destination_cidrs.0", "0.0.0.0/0"),
				),
			},
			{
				Config: `data "gamefabric_protection_gatewaypolicy" "test2" {
  display_name = "My Gateway Policy"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test2", "name", "test-gateway-policy"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test2", "display_name", "My Gateway Policy"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test2", "description", "My Gateway Policy Description"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test2", "destination_cidrs.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_gatewaypolicy.test2", "destination_cidrs.0", "0.0.0.0/0"),
				),
			},
		},
	})
}

func TestGatewayPolicy_HandlesMultipleMatches(t *testing.T) {
	pol := &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test-gateway-policy"},
		Spec: protectionv1.GatewayPolicySpec{
			DisplayName: "My Gateway Policy",
		},
	}
	otherPol := &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "other"},
		Spec: protectionv1.GatewayPolicySpec{
			DisplayName: "My Gateway Policy",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, pol, otherPol)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_gatewaypolicy" "test" {
  display_name = "My Gateway Policy"
}
`,
				ExpectError: regexp.MustCompile(`Multiple Gateway Policies Found`),
			},
		},
	})
}

func TestGatewayPolicy_HandlesMultipleSelectors(t *testing.T) {
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_gatewaypolicy" "test" {
  name = "test-gateway-policy"
  display_name = "My Gateway Policy"
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
