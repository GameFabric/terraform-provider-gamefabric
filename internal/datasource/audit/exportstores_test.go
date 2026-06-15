package audit_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestExportStores(t *testing.T) {
	t.Parallel()

	es1 := &auditv1alpha1.ExportStore{
		ObjectMeta: metav1.ObjectMeta{Name: "es-1"},
		Spec: auditv1alpha1.ExportStoreSpec{
			S3: &auditv1alpha1.ExportStoreS3{
				Bucket: "bucket-1",
				Region: "eu-west-1",
				Auth: auditv1alpha1.ExportStoreS3Auth{
					AccessKeyID:     "key-1",
					SecretAccessKey: auditv1alpha1.MaskedValue,
				},
			},
		},
		Status: auditv1alpha1.ExportStoreStatus{
			State: auditv1alpha1.ExportStoreStateActive,
		},
	}
	es2 := &auditv1alpha1.ExportStore{
		ObjectMeta: metav1.ObjectMeta{Name: "es-2"},
		Spec: auditv1alpha1.ExportStoreSpec{
			Suspended: true,
			S3: &auditv1alpha1.ExportStoreS3{
				Bucket: "bucket-2",
				Region: "us-east-1",
				Auth: auditv1alpha1.ExportStoreS3Auth{
					AccessKeyID:     "key-2",
					SecretAccessKey: auditv1alpha1.MaskedValue,
				},
			},
		},
		Status: auditv1alpha1.ExportStoreStatus{
			State: auditv1alpha1.ExportStoreStateSuspended,
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, es1, es2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_exportstores" "all" { }
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.0.name", "es-1"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.0.suspended", "false"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.0.state", "Active"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.1.name", "es-2"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.1.suspended", "true"),
					resource.TestCheckResourceAttr("data.gamefabric_exportstores.all", "exportstores.1.state", "Suspended"),
				),
			},
		},
	})
}
