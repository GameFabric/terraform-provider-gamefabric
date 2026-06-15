package audit

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type exportStoreModel struct {
	ID        types.String        `tfsdk:"id"`
	Name      types.String        `tfsdk:"name"`
	Suspended types.Bool          `tfsdk:"suspended"`
	S3        *exportStoreS3Model `tfsdk:"s3"`
}

type exportStoreS3Model struct {
	Bucket   types.String           `tfsdk:"bucket"`
	Region   types.String           `tfsdk:"region"`
	Endpoint types.String           `tfsdk:"endpoint"`
	Prefix   types.String           `tfsdk:"prefix"`
	Auth     exportStoreS3AuthModel `tfsdk:"auth"`
}

type exportStoreS3AuthModel struct {
	AccessKeyID            types.String `tfsdk:"access_key_id"`
	SecretAccessKey        types.String `tfsdk:"secret_access_key"`
	SecretAccessKeyVersion types.Int64  `tfsdk:"secret_access_key_version"`
}

func newExportStoreModel(obj *auditv1alpha1.ExportStore) exportStoreModel {
	m := exportStoreModel{
		ID:        types.StringValue(obj.Name),
		Name:      types.StringValue(obj.Name),
		Suspended: types.BoolValue(obj.Spec.Suspended),
	}

	if obj.Spec.S3 != nil {
		m.S3 = &exportStoreS3Model{
			Bucket:   types.StringValue(obj.Spec.S3.Bucket),
			Region:   types.StringValue(obj.Spec.S3.Region),
			Endpoint: types.StringValue(obj.Spec.S3.Endpoint),
			Prefix:   types.StringValue(obj.Spec.S3.Prefix),
			Auth: exportStoreS3AuthModel{
				AccessKeyID: types.StringValue(obj.Spec.S3.Auth.AccessKeyID),
				// SecretAccessKey is write-only: never populated from the API response.
				SecretAccessKey:        types.StringNull(),
				SecretAccessKeyVersion: types.Int64Null(),
			},
		}
	}

	return m
}

func (m exportStoreModel) ToObject() *auditv1alpha1.ExportStore {
	obj := &auditv1alpha1.ExportStore{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name.ValueString(),
		},
		Spec: auditv1alpha1.ExportStoreSpec{
			Suspended: m.Suspended.ValueBool(),
		},
	}

	if m.S3 != nil {
		secretKey := m.S3.Auth.SecretAccessKey.ValueString()
		if secretKey == "" {
			// Write-only field is null in state: send the masked sentinel so the
			// server's BeforeUpdate transparently restores the stored credential.
			secretKey = auditv1alpha1.MaskedValue
		}

		obj.Spec.S3 = &auditv1alpha1.ExportStoreS3{
			Bucket:   m.S3.Bucket.ValueString(),
			Region:   m.S3.Region.ValueString(),
			Endpoint: m.S3.Endpoint.ValueString(),
			Prefix:   m.S3.Prefix.ValueString(),
			Auth: auditv1alpha1.ExportStoreS3Auth{
				AccessKeyID:     m.S3.Auth.AccessKeyID.ValueString(),
				SecretAccessKey: secretKey,
			},
		}
	}

	return obj
}
