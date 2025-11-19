package storage_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/storage"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators/validatorstest"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestVolume(t *testing.T) {
	t.Parallel()

	name := "eu"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceVolumeDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceVolumeConfigBasic(name, "10M"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_volume.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "volume_store", "test-volume-store"),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "capacity", "10M"),
				),
			},
			{
				ResourceName:      "gamefabric_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceVolumeConfigBasicWithLabels(name, "10M"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_volume.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "labels.foo", "bar"),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "volume_store", "test-volume-store"),
					resource.TestCheckResourceAttr("gamefabric_volume.test", "capacity", "10M"),
				),
			},
			{
				Config:      testResourceVolumeConfigBasicWithLabels(name, "8M"),
				ExpectError: regexp.MustCompile(`The capacity of the volume .+ cannot be decreased`),
			},
		},
	})
}

func TestVolumeResourceGameFabricValidators(t *testing.T) {
	t.Parallel()

	resp := &tfresource.SchemaResponse{}

	volume := storage.NewVolume()
	volume.Schema(t.Context(), tfresource.SchemaRequest{}, resp)

	want := validatorstest.CollectJSONPaths(&storagev1beta1.Volume{})
	got := validatorstest.CollectPathExpressions(resp.Schema)

	require.NotEmpty(t, got)
	require.NotEmpty(t, want)
	for _, path := range got {
		require.Containsf(t, want, path, "The validator path %q was not found in the Volume API object", path)
	}
}

func testResourceVolumeConfigBasic(name, size string) string {
	return fmt.Sprintf(`resource "gamefabric_volume" "test" {
  name = "%s"
  environment = "dflt"
  volume_store = "test-volume-store"
  capacity = "%s"
}`, name, size)
}

func testResourceVolumeConfigBasicWithLabels(name, size string) string {
	return fmt.Sprintf(`resource "gamefabric_volume" "test" {
  name = "%s"
  environment = "dflt"
  labels = {
    foo = "bar"
  }
  volume_store = "test-volume-store"
  capacity = "%s"
}`, name, size)
}

func testResourceVolumeDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_volume" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.StorageV1Beta1().Volumes(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil && resp.Name == name {
				return fmt.Errorf("volume still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
