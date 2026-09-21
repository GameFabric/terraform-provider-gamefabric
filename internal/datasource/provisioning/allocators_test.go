package provisioning_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAllocatorsDataSource(t *testing.T) {
	a := &provisioningv1beta1.Allocator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "allocator-a",
			Labels: map[string]string{"region": "eu-west"},
		},
		Spec: provisioningv1beta1.AllocatorSpec{
			Region: "eu-west",
		},
	}
	b := &provisioningv1beta1.Allocator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "allocator-b",
			Labels: map[string]string{"region": "us-east"},
		},
		Spec: provisioningv1beta1.AllocatorSpec{
			Region: "us-east",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, a, b)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_allocators" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_allocators.all", "allocators.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_allocators.all", "allocators.0.name", "allocator-a"),
					resource.TestCheckResourceAttr("data.gamefabric_allocators.all", "allocators.0.region", "eu-west"),
					resource.TestCheckResourceAttr("data.gamefabric_allocators.all", "allocators.1.name", "allocator-b"),
					resource.TestCheckResourceAttr("data.gamefabric_allocators.all", "allocators.1.region", "us-east"),
				),
			},
			{
				Config: `data "gamefabric_allocators" "filtered" {
  label_filter = {
    region = "eu-west"
  }
}
`,
				// The fake client set does not implement server-side label selector filtering
				// (FakeClientSet.List ignores metav1.ListOptions entirely), so this only verifies
				// that label_filter round-trips into state correctly. Actual filtering is
				// delegated to and covered by GCAP's server-side implementation.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_allocators.filtered", "label_filter.region", "eu-west"),
				),
			},
		},
	})
}

func TestAllocatorsDataSource_Empty(t *testing.T) {
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_allocators" "all" {}
`,
				// allocators must be an empty list, not null, so it can safely be used
				// with collection operations like length() or for_each.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_allocators.all", "allocators.#", "0"),
				),
			},
		},
	})
}
