package audit_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestExportStore(t *testing.T) {
	t.Parallel()

	es := &auditv1alpha1.ExportStore{
		ObjectMeta: metav1.ObjectMeta{Name: "test-exportstore"},
		Spec: auditv1alpha1.ExportStoreSpec{
			S3: &auditv1alpha1.ExportStoreS3{
				Endpoint: "https://s3.amazonaws.com",
				Region:   "eu-west-1",
				Bucket:   "my-audit-bucket",
				Prefix:   "logs/",
				Auth: auditv1alpha1.ExportStoreS3Auth{
					AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
					SecretAccessKey: auditv1alpha1.MaskedValue,
				},
			},
		},
		Status: auditv1alpha1.ExportStoreStatus{
			State:   auditv1alpha1.ExportStoreStateActive,
			Message: "Delivering successfully",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, es)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_exportstore" "test" {
  name = "test-exportstore"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "name", "test-exportstore"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "suspended", "false"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "s3.endpoint", "https://s3.amazonaws.com"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "s3.region", "eu-west-1"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "s3.bucket", "my-audit-bucket"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "s3.prefix", "logs/"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "s3.auth.access_key_id", "AKIAIOSFODNN7EXAMPLE"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "s3.auth.secret_access_key", auditv1alpha1.MaskedValue),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "state", "Active"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstore.test", "message", "Delivering successfully"),
				),
			},
		},
	})
}

func TestExportStore_NotFound(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_exportstore" "missing" {
  name = "does-not-exist"
}
`,
				ExpectError: regexp.MustCompile(`ExportStore Not Found`),
			},
		},
	})
}
