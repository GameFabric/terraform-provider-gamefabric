package audit_test

import (
	"fmt"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestExportStoreResource(t *testing.T) {
	t.Parallel()

	name := "test-store"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceExportStoreDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceExportStoreConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "suspended", "false"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.bucket", "my-audit-bucket"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.region", "eu-west-1"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.access_key_id", "AKIAIOSFODNN7EXAMPLE"),
					// secret_access_key is write-only and must never appear in state.
					resource.TestCheckNoResourceAttr("gamefabric_exportstore.test", "s3.auth.secret_access_key"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.secret_access_key_version", "1"),
				),
			},
			{
				ResourceName:      "gamefabric_exportstore.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Write-only fields and optional-computed s3 fields cannot be verified
				// on import as the fake API does not run BeforeCreate hooks.
				ImportStateVerifyIgnore: []string{
					"s3.auth.secret_access_key",
					"s3.auth.secret_access_key_version",
					"s3.endpoint",
					"s3.prefix",
					"suspended",
				},
			},
			{
				Config: testResourceExportStoreConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "suspended", "false"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.bucket", "my-audit-bucket"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.region", "eu-west-1"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.prefix", "audit/"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.access_key_id", "AKIAIOSFODNN7EXAMPLE"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.secret_access_key_version", "1"),
				),
			},
		},
	})
}

func TestExportStoreResource_Suspended(t *testing.T) {
	t.Parallel()

	name := "test-store-suspended"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceExportStoreDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceExportStoreConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "suspended", "false"),
				),
			},
			{
				Config: testResourceExportStoreConfigSuspended(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "suspended", "true"),
				),
			},
		},
	})
}

func TestExportStoreResource_KeyRotation(t *testing.T) {
	t.Parallel()

	name := "test-store-rotation"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceExportStoreDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceExportStoreConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.secret_access_key_version", "1"),
				),
			},
			{
				Config: testResourceExportStoreConfigRotatedKey(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.access_key_id", "AKIAIOSFODNN7EXAMPLE"),
					resource.TestCheckResourceAttr("gamefabric_exportstore.test", "s3.auth.secret_access_key_version", "2"),
				),
			},
		},
	})
}

func testResourceExportStoreConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_exportstore" "test" {
  name = "%s"

  s3 = {
    bucket = "my-audit-bucket"
    region = "eu-west-1"

    auth = {
      access_key_id             = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      secret_access_key_version = 1
    }
  }
}`, name)
}

func testResourceExportStoreConfigUpdated(name string) string {
	return fmt.Sprintf(`resource "gamefabric_exportstore" "test" {
  name = "%s"

  s3 = {
    bucket = "my-audit-bucket"
    region = "eu-west-1"
    prefix = "audit/"

    auth = {
      access_key_id             = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      secret_access_key_version = 1
    }
  }
}`, name)
}

func testResourceExportStoreConfigSuspended(name string) string {
	return fmt.Sprintf(`resource "gamefabric_exportstore" "test" {
  name      = "%s"
  suspended = true

  s3 = {
    bucket = "my-audit-bucket"
    region = "eu-west-1"

    auth = {
      access_key_id             = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      secret_access_key_version = 1
    }
  }
}`, name)
}

func testResourceExportStoreConfigRotatedKey(name string) string {
	return fmt.Sprintf(`resource "gamefabric_exportstore" "test" {
  name = "%s"

  s3 = {
    bucket = "my-audit-bucket"
    region = "eu-west-1"

    auth = {
      access_key_id             = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key         = "newSecretKeyAfterRotation"
      secret_access_key_version = 2
    }
  }
}`, name)
}

func testResourceExportStoreDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_exportstore" {
				continue
			}

			resp, err := cs.AuditV1Alpha1().ExportStores().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil && resp.Name == rs.Primary.ID {
				return fmt.Errorf("exportstore still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
