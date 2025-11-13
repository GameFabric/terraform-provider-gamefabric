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

func TestVolumeStores(t *testing.T) {
	t.Parallel()

	vs1 := &storagev1.VolumeStore{
		ObjectMeta: metav1.ObjectMeta{Name: "vs-1"},
		Spec: storagev1.VolumeStoreSpec{
			Region:        "eu-central1",
			Destination:   "gs://b1",
			MaxVolumeSize: *conv.Quantity(typesframework.StringValue("10Mi")),
		},
	}
	vs2 := &storagev1.VolumeStore{
		ObjectMeta: metav1.ObjectMeta{Name: "vs-2"},
		Spec: storagev1.VolumeStoreSpec{
			Region:        "us-west1",
			Destination:   "gs://b2",
			MaxVolumeSize: *conv.Quantity(typesframework.StringValue("20Mi")),
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vs1, vs2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volumestores" "all" { }
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_volumestores.all", "volumestores.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_volumestores.all", "volumestores.0.name", "vs-1"),
					resource.TestCheckResourceAttr("data.gamefabric_volumestores.all", "volumestores.1.name", "vs-2"),
				),
			},
		},
	})
}
