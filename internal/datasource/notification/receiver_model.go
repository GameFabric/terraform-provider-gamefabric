package notification

import (
	notificationv1alpha1 "github.com/gamefabric/gf-core/pkg/api/notification/v1alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type receiverModel struct {
	Name        types.String            `tfsdk:"name"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Annotations map[string]types.String `tfsdk:"annotations"`
	EmailTo     []types.String          `tfsdk:"email_to"`
}

func newReceiverModel(obj *notificationv1alpha1.Receiver) receiverModel {
	var emailTo []string
	if obj.Spec.Email != nil {
		emailTo = obj.Spec.Email.To
	}

	return receiverModel{
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		EmailTo:     conv.ForEachSliceItem(emailTo, func(item string) types.String { return types.StringValue(item) }),
	}
}
