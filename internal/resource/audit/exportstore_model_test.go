package audit

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewExportStoreModel(t *testing.T) {
	model := newExportStoreModel(testExportStoreObject)
	assert.Equal(t, testExportStoreModel, model)
}

func TestExportStoreModel_ToObject(t *testing.T) {
	obj := testExportStoreModel.ToObject()
	assert.Equal(t, testExportStoreObject.Name, obj.Name)
	assert.Equal(t, testExportStoreObject.Spec.Suspended, obj.Spec.Suspended)
	assert.Equal(t, testExportStoreObject.Spec.S3.Bucket, obj.Spec.S3.Bucket)
	assert.Equal(t, testExportStoreObject.Spec.S3.Region, obj.Spec.S3.Region)
	assert.Equal(t, testExportStoreObject.Spec.S3.Endpoint, obj.Spec.S3.Endpoint)
	assert.Equal(t, testExportStoreObject.Spec.S3.Prefix, obj.Spec.S3.Prefix)
	assert.Equal(t, testExportStoreObject.Spec.S3.Auth.AccessKeyID, obj.Spec.S3.Auth.AccessKeyID)
	// SecretAccessKey is null in the model (write-only): ToObject emits the masked sentinel.
	assert.Equal(t, auditv1alpha1.MaskedValue, obj.Spec.S3.Auth.SecretAccessKey)
}

var (
	testExportStoreObject = &auditv1alpha1.ExportStore{
		ObjectMeta: metav1.ObjectMeta{Name: "my-store"},
		Spec: auditv1alpha1.ExportStoreSpec{
			S3: &auditv1alpha1.ExportStoreS3{
				Bucket:   "my-bucket",
				Region:   "eu-west-1",
				Endpoint: "https://s3.amazonaws.com",
				Prefix:   "logs/",
				Auth: auditv1alpha1.ExportStoreS3Auth{
					AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
					SecretAccessKey: auditv1alpha1.MaskedValue,
				},
			},
		},
	}

	testExportStoreModel = exportStoreModel{
		ID:        types.StringValue("my-store"),
		Name:      types.StringValue("my-store"),
		Suspended: types.BoolValue(false),
		S3: &exportStoreS3Model{
			Bucket:   types.StringValue("my-bucket"),
			Region:   types.StringValue("eu-west-1"),
			Endpoint: types.StringValue("https://s3.amazonaws.com"),
			Prefix:   types.StringValue("logs/"),
			Auth: exportStoreS3AuthModel{
				AccessKeyID:            types.StringValue("AKIAIOSFODNN7EXAMPLE"),
				SecretAccessKey:        types.StringNull(),
				SecretAccessKeyVersion: types.Int64Null(),
			},
		},
	}
)
