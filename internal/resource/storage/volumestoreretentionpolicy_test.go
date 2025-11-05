package storage_test

import (
	"fmt"
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

func TestVolumeStoreRetentionPolicy(t *testing.T) {
	t.Parallel()

	name := "test-retention-policy"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceVolumeStoreRetentionPolicyDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceVolumeStoreRetentionPolicyConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "volume_store", "default-volumestore"),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "offline_snapshot.count", "12"),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "offline_snapshot.days", "30"),
				),
			},
			{
				ResourceName: "gamefabric_volumestore_retention_policy.test",
				ImportState:  true,
			},
			{
				Config: testResourceVolumeStoreRetentionPolicyConfigBasicWithOnline(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "volume_store", "default-volumestore"),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "online_snapshot.count", "12"),
					resource.TestCheckResourceAttr("gamefabric_volumestore_retention_policy.test", "online_snapshot.days", "30"),
				),
			},
		},
	})
}

func TestVolumeStoreRetentionPolicyResourceGameFabricValidators(t *testing.T) {
	t.Parallel()

	resp := &tfresource.SchemaResponse{}

	vsrp := storage.NewVolumeStoreRetentionPolicy()
	vsrp.Schema(t.Context(), tfresource.SchemaRequest{}, resp)

	want := validatorstest.CollectJSONPaths(&storagev1beta1.VolumeStoreRetentionPolicy{})
	got := validatorstest.CollectPathExpressions(resp.Schema)

	require.NotEmpty(t, got)
	require.NotEmpty(t, want)
	for _, path := range got {
		require.Containsf(t, want, path, "The validator path %q was not found in the Volume Store Retention Policy API object", path)
	}
}

func testResourceVolumeStoreRetentionPolicyConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_volumestore_retention_policy" "test" {
  name = "%s"
  environment = "dflt"
  volume_store = "default-volumestore"
  offline_snapshot = {
    count = 12
    days = 30
  }
}`, name)
}

func testResourceVolumeStoreRetentionPolicyConfigBasicWithOnline(name string) string {
	return fmt.Sprintf(`resource "gamefabric_volumestore_retention_policy" "test" {
  name = "%s"
  environment = "dflt"
  volume_store = "default-volumestore"
  online_snapshot = {
    count = 12
    days = 30
  }
}`, name)
}

func testResourceVolumeStoreRetentionPolicyDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_volumestore_retention_policy" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.StorageV1Beta1().VolumeStoreRetentionPolicies(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil {
				if resp.Name == name {
					return fmt.Errorf("volume store retention policy still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
