package storage_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVolumes(t *testing.T) {
	t.Parallel()

	vol := &storagev1beta1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-volume",
			Environment: "dflt",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vol)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volumes" "test1" {
  environment = "dflt"
  label_filter = {
    foo = "bar"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "label_filter.foo", "bar"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.0.name", "test-volume"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.0.environment", "dflt"),
				),
			},
		},
	})
}

func TestVolumes_AllowsGettingAll(t *testing.T) {
	t.Parallel()

	vol1 := &storagev1beta1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-volume1",
			Environment: "dflt",
		},
	}
	vol2 := &storagev1beta1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-volume2",
			Environment: "dflt",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, vol1, vol2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_volumes" "test1" {
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.0.name", "test-volume1"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.0.environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.1.name", "test-volume2"),
					resource.TestCheckResourceAttr("data.gamefabric_volumes.test1", "volumes.1.environment", "dflt"),
				),
			},
		},
	})
}
