package protection

import (
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
)

type protocolsModel struct {
	Protocols []protocolModel `tfsdk:"protocols"`
}

func newProtocolsModel(objs []protectionv1.Protocol) protocolsModel {
	return protocolsModel{
		Protocols: conv.ForEachSliceItem(objs, func(item protectionv1.Protocol) protocolModel {
			return newProtocolModel(&item)
		}),
	}
}
