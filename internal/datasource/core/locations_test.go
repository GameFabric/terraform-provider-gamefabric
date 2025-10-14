package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLocations(t *testing.T) {
	t.Parallel()

	loc1 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-1",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc2 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc2",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc3 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-3",
			Annotations: map[string]string{
				"g8c.io/provider-type": "something",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc4 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-4",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "something",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc5 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-5",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "something",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc6 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-6",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "something",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, loc1, loc2, loc3, loc4, loc5, loc6)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  type = "Baremetal"
  city = "Amsterdam"
  country = "Netherlands"
  continent = "Europe"
  name_regex = "loc-.*"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "type", "Baremetal"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "city", "Amsterdam"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "country", "Netherlands"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "continent", "Europe"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "name_regex", "loc-.*"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "locations.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "locations.0", "loc-1"),
				),
			},
		},
	})
}
