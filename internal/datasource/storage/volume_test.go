package storage_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVolume(t *testing.T) {
	t.Parallel()

	vol := &storagev1beta1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-volume",
			Environment: "dflt",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vol)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volume" "test1" {
  name = "test-volume"
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_volume.test1", "name", "test-volume"),
					resource.TestCheckResourceAttr("data.gamefabric_volume.test1", "environment", "dflt"),
				),
			},
		},
	})
}
