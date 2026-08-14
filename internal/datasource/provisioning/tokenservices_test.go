package provisioning_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTokenServicesDataSource(t *testing.T) {
	a := &provisioningv1beta1.TokenService{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "service-a",
			Labels: map[string]string{"team": "platforms"},
		},
		Spec: provisioningv1beta1.TokenServiceSpec{
			Environment: provisioningv1beta1.TokenServiceEnvProd,
			Game:        provisioningv1beta1.TokenServiceGameSpec{Name: "game-a"},
			JWKS:        &provisioningv1beta1.TokenServiceJWKSSpec{URL: "https://a.example.com/jwks.json"},
		},
	}
	b := &provisioningv1beta1.TokenService{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "service-b",
			Labels: map[string]string{"team": "other"},
		},
		Spec: provisioningv1beta1.TokenServiceSpec{
			Environment: provisioningv1beta1.TokenServiceEnvDev,
			Game:        provisioningv1beta1.TokenServiceGameSpec{Name: "game-b"},
			JWKS:        &provisioningv1beta1.TokenServiceJWKSSpec{URL: "https://b.example.com/jwks.json"},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, a, b)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_steelshield_tokenservices" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.all", "token_services.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.all", "token_services.0.name", "service-a"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.all", "token_services.1.name", "service-b"),
					// service-a is prod (development_mode=false), service-b is dev (development_mode=true).
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.all", "token_services.0.development_mode", "false"),
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.all", "token_services.1.development_mode", "true"),
				),
			},
			{
				Config: `data "gamefabric_steelshield_tokenservices" "filtered" {
  label_filter = {
    team = "platforms"
  }
}
`,
				// The fake client set does not implement server-side label selector filtering
				// (FakeClientSet.List ignores metav1.ListOptions entirely), so this only verifies
				// that label_filter round-trips into state correctly. Actual filtering is
				// delegated to and covered by GCAP's server-side implementation.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.filtered", "label_filter.team", "platforms"),
				),
			},
		},
	})
}

func TestTokenServicesDataSource_Empty(t *testing.T) {
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_steelshield_tokenservices" "all" {}
`,
				// token_services must be an empty list, not null, so it can safely be used
				// with collection operations like length() or for_each.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_steelshield_tokenservices.all", "token_services.#", "0"),
				),
			},
		},
	})
}
