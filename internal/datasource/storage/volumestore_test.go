package storage_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	typesframework "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVolumeStore_ByName(t *testing.T) {
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

func TestVolumeStore_ByRegion(t *testing.T) {
	t.Parallel()

	vs := &storagev1.VolumeStore{
		ObjectMeta: metav1.ObjectMeta{Name: "vs-region"},
		Spec: storagev1.VolumeStoreSpec{
			Region:        "europe-west1",
			Destination:   "gs://bucket2",
			MaxVolumeSize: *conv.Quantity(typesframework.StringValue("5Gi")),
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vs)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volumestore" "by_region" {
  region = "europe-west1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_volumestore.by_region", "name", "vs-region"),
					resource.TestCheckResourceAttr("data.gamefabric_volumestore.by_region", "region", "europe-west1"),
					resource.TestCheckResourceAttr("data.gamefabric_volumestore.by_region", "max_volume_size", "5Gi"),
				),
			},
		},
	})
}

func TestVolumeStore_ByRegion_MultipleMatches(t *testing.T) {
	t.Parallel()

	vs1 := &storagev1.VolumeStore{
		ObjectMeta: metav1.ObjectMeta{Name: "vs-a"},
		Spec: storagev1.VolumeStoreSpec{
			Region:        "dup-region",
			Destination:   "gs://b1",
			MaxVolumeSize: *conv.Quantity(typesframework.StringValue("1Gi")),
		},
	}
	vs2 := &storagev1.VolumeStore{
		ObjectMeta: metav1.ObjectMeta{Name: "vs-b"},
		Spec: storagev1.VolumeStoreSpec{
			Region:        "dup-region",
			Destination:   "gs://b2",
			MaxVolumeSize: *conv.Quantity(typesframework.StringValue("2Gi")),
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vs1, vs2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volumestore" "dup" {
  region = "dup-region"
}
`,
				ExpectError: regexp.MustCompile(`Multiple VolumeStores Found`),
			},
		},
	})
}
