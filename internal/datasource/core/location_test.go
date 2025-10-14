package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLocation(t *testing.T) {
	t.Parallel()

	loc := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{Name: "test-location"},
		Spec: corev1.LocationSpec{
			Sites: []string{"test-site-1", "test-site-2"},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, loc)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_location" "test1" {
  name = "test-location"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_location.test1", "name", "test-location"),
				),
			},
		},
	})
}
