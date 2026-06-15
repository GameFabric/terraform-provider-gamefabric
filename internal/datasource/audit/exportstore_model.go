package audit

import (
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type exportStoreModel struct {
	Name      types.String        `tfsdk:"name"`
	Suspended types.Bool          `tfsdk:"suspended"`
	S3        *exportStoreS3Model `tfsdk:"s3"`
	State     types.String        `tfsdk:"state"`
	Message   types.String        `tfsdk:"message"`
}

type exportStoreS3Model struct {
	Endpoint types.String           `tfsdk:"endpoint"`
	Region   types.String           `tfsdk:"region"`
	Bucket   types.String           `tfsdk:"bucket"`
	Prefix   types.String           `tfsdk:"prefix"`
	Auth     exportStoreS3AuthModel `tfsdk:"auth"`
}

type exportStoreS3AuthModel struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

func newExportStoreModel(obj *auditv1alpha1.ExportStore) exportStoreModel {
	m := exportStoreModel{
		Name:      types.StringValue(obj.Name),
		Suspended: types.BoolValue(obj.Spec.Suspended),
		State:     types.StringValue(string(obj.Status.State)),
		Message:   conv.OptionalFunc(obj.Status.Message, types.StringValue, types.StringNull),
	}

	if obj.Spec.S3 != nil {
		s3 := &exportStoreS3Model{
			Endpoint: conv.OptionalFunc(obj.Spec.S3.Endpoint, types.StringValue, types.StringNull),
			Region:   conv.OptionalFunc(obj.Spec.S3.Region, types.StringValue, types.StringNull),
			Bucket:   types.StringValue(obj.Spec.S3.Bucket),
			Prefix:   conv.OptionalFunc(obj.Spec.S3.Prefix, types.StringValue, types.StringNull),
			Auth: exportStoreS3AuthModel{
				AccessKeyID:     types.StringValue(obj.Spec.S3.Auth.AccessKeyID),
				SecretAccessKey: types.StringValue(obj.Spec.S3.Auth.SecretAccessKey),
			},
		}
		m.S3 = s3
	}

	return m
}
