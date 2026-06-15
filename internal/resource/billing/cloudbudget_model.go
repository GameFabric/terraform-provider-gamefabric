package billing

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	billingv2alpha1 "github.com/gamefabric/gf-core/pkg/api/billing/v2alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type cloudBudgetModel struct {
	ID          types.String              `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	Labels      map[string]types.String   `tfsdk:"labels"`
	Annotations map[string]types.String   `tfsdk:"annotations"`
	Suspended   types.Bool                `tfsdk:"suspended"`
	Receivers   []types.String            `tfsdk:"receivers"`
	MaxBudget   types.Float64             `tfsdk:"max_budget"`
	Thresholds  []types.String            `tfsdk:"thresholds"`
	Interval    *cloudBudgetIntervalModel `tfsdk:"interval"`
}

type cloudBudgetIntervalModel struct {
	Start types.String `tfsdk:"start"`
	Step  types.String `tfsdk:"step"`
}

func newCloudBudgetModel(obj *billingv2alpha1.CloudBudget) cloudBudgetModel {
	var interval *cloudBudgetIntervalModel
	if obj.Spec.Interval != nil {
		interval = &cloudBudgetIntervalModel{
			Start: conv.FromIntOrString(&obj.Spec.Interval.Start),
			Step:  conv.FromIntOrString(&obj.Spec.Interval.Step),
		}
	}

	return cloudBudgetModel{
		ID:          types.StringValue(obj.Name),
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		Suspended:   types.BoolValue(obj.Spec.Suspended),
		Receivers:   conv.ForEachSliceItem(obj.Spec.Receivers, func(item string) types.String { return types.StringValue(item) }),
		MaxBudget:   types.Float64Value(obj.Spec.MaxBudget),
		Thresholds:  conv.ForEachSliceItem(obj.Spec.Thresholds, func(item intstr.IntOrString) types.String { return conv.FromIntOrString(&item) }),
		Interval:    interval,
	}
}

func (m cloudBudgetModel) ToObject() *billingv2alpha1.CloudBudget {
	var interval *billingv2alpha1.CloudBudgetInterval
	if m.Interval != nil {
		start := intstr.Parse(m.Interval.Start.ValueString())
		step := intstr.Parse(m.Interval.Step.ValueString())
		interval = &billingv2alpha1.CloudBudgetInterval{
			Start: start,
			Step:  step,
		}
	}

	return &billingv2alpha1.CloudBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: billingv2alpha1.CloudBudgetSpec{
			Suspended: m.Suspended.ValueBool(),
			Receivers: conv.ForEachSliceItem(m.Receivers, func(v types.String) string { return v.ValueString() }),
			MaxBudget: m.MaxBudget.ValueFloat64(),
			Thresholds: conv.ForEachSliceItem(m.Thresholds, func(v types.String) intstr.IntOrString {
				return intstr.Parse(v.ValueString())
			}),
			Interval: interval,
		},
	}
}
