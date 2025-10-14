package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	kresource "k8s.io/apimachinery/pkg/api/resource"
)

func TestRegions(t *testing.T) {
	region1 := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "region-1",
			Environment: "dflt",
			Labels: map[string]string{
				"baremetal": "true",
			},
		},
		Spec: corev1.RegionSpec{
			Types: []corev1.RegionType{
				{
					Name: "baremetal",
				},
			},
		},
		Status: corev1.RegionStatus{
			Types: []corev1.RegionTypeStatus{
				{
					Name: "baremetal",
					Used: corev1.Resources{
						CPU:    kresource.MustParse("4"),
						Memory: kresource.MustParse("8Gi"),
					},
					Limit: corev1.Resources{
						CPU:    kresource.MustParse("6"),
						Memory: kresource.MustParse("12Gi"),
					},
				},
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, region1)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_regions" "test1" {
  environment = "dflt"
  label_filter = {
    baremetal = "true"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_regions.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_regions.test1", "label_filter.baremetal", "true"),
					resource.TestCheckResourceAttr("data.gamefabric_regions.test1", "regions.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_regions.test1", "regions.0.types.baremetal.cpu", "6"),
					resource.TestCheckResourceAttr("data.gamefabric_regions.test1", "regions.0.types.baremetal.memory", "12Gi"),
				),
			},
		},
	})
}
