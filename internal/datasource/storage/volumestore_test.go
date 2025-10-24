package storage_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	typesframework "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVolumeStore(t *testing.T) {
	t.Parallel()

	vs := &storagev1.VolumeStore{
		ObjectMeta: metav1.ObjectMeta{Name: "test-volumestore"},
		Spec: storagev1.VolumeStoreSpec{
			Region:        "europe-west1",
			Destination:   "gs://bucket",
			MaxVolumeSize: *conv.Quantity(typesframework.StringValue("10Mi")),
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vs)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volumestore" "test" {
  name = "test-volumestore"
  region = "europe-west1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_volumestore.test", "name", "test-volumestore"),
					resource.TestCheckResourceAttr("data.gamefabric_volumestore.test", "region", "europe-west1"),
					resource.TestCheckResourceAttr("data.gamefabric_volumestore.test", "max_volume_size", "10Mi"),
				),
			},
		},
	})
}
