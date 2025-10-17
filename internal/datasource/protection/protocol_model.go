package protection

import (
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type protocolModel struct {
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Protocol    types.String `tfsdk:"protocol"`
}

func newProtocolModel(obj *protectionv1.Protocol) protocolModel {
	model := protocolModel{
		Name:        conv.OptionalFunc(obj.Name, types.StringValue, types.StringNull),
		DisplayName: conv.OptionalFunc(obj.Spec.DisplayName, types.StringValue, types.StringNull),
		Description: conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Protocol:    conv.OptionalFunc(string(obj.Spec.Protocol), types.StringValue, types.StringNull),
	}
	return model
}
