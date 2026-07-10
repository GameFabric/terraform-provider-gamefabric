package provisioning_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAllocator(t *testing.T) {
	t.Parallel()

	alloc := &provisioningv1beta1.Allocator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-allocator",
		},
		Spec: provisioningv1beta1.AllocatorSpec{
			Region: "eu-west",
			RateLimit: provisioningv1beta1.RateLimit{
				QPS:   100,
				Burst: 200,
			},
		},
		Status: provisioningv1beta1.AllocatorStatus{
			Allocation: provisioningv1beta1.AllocatorEndpoint{
				URL:    "https://alloc.example.com",
				Tokens: []string{"alloc-token-old", "alloc-token-new"},
			},
			Registration: provisioningv1beta1.AllocatorEndpoint{
				URL:    "https://reg.example.com",
				Tokens: []string{"reg-token-old", "reg-token-new"},
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, alloc)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_allocator" "test" {
  name = "test-allocator"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "name", "test-allocator"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "region", "eu-west"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "rate_limit_qps", "100"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "rate_limit_burst", "200"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "allocation_url", "https://alloc.example.com"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "allocation_active_token", "alloc-token-new"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "allocation_tokens.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "allocation_tokens.0", "alloc-token-old"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "allocation_tokens.1", "alloc-token-new"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "registry_url", "https://reg.example.com"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "registry_active_token", "reg-token-new"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "registry_tokens.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "registry_tokens.0", "reg-token-old"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "registry_tokens.1", "reg-token-new"),
				),
			},
		},
	})
}

func TestAllocator_NoRateLimit(t *testing.T) {
	t.Parallel()

	alloc := &provisioningv1beta1.Allocator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "minimal-allocator",
		},
		Spec: provisioningv1beta1.AllocatorSpec{
			Region: "us-east",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, alloc)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_allocator" "test" {
  name = "minimal-allocator"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "name", "minimal-allocator"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "region", "us-east"),
					resource.TestCheckNoResourceAttr("data.gamefabric_allocator.test", "rate_limit_qps"),
					resource.TestCheckNoResourceAttr("data.gamefabric_allocator.test", "rate_limit_burst"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "allocation_tokens.#", "0"),
					resource.TestCheckResourceAttr("data.gamefabric_allocator.test", "registration_tokens.#", "0"),
				),
			},
		},
	})
}
