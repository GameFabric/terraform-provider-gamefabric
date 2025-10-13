package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLocations(t *testing.T) {
	loc1 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-1",
			Labels: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/continent":     "eu",
			},
		},
	}
	loc2 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc2",
			Labels: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/continent":     "eu",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, loc1, loc2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  type = "baremetal"
  city = "amsterdam"
  continent = "eu"
  name_regex = "loc-.*"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "type", "baremetal"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "city", "amsterdam"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "continent", "eu"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "name_regex", "loc-.*"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "locations.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "locations.0", "loc-1"),
				),
			},
		},
	})
}
