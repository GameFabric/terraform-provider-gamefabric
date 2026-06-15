package notification

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	notificationv1alpha1 "github.com/gamefabric/gf-core/pkg/api/notification/v1alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type receiverModel struct {
	ID          types.String            `tfsdk:"id"`
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
		ID:          types.StringValue(obj.Name),
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		EmailTo:     conv.ForEachSliceItem(emailTo, func(item string) types.String { return types.StringValue(item) }),
	}
}

func (m receiverModel) ToObject() *notificationv1alpha1.Receiver {
	return &notificationv1alpha1.Receiver{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: notificationv1alpha1.ReceiverSpec{
			Email: &notificationv1alpha1.ReceiverEmail{
				To: conv.ForEachSliceItem(m.EmailTo, func(v types.String) string { return v.ValueString() }),
			},
		},
	}
}
