package container

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type branchModel struct {
	ID                   types.String                          `tfsdk:"id"`
	Name                 types.String                          `tfsdk:"name"`
	Labels               map[string]types.String               `tfsdk:"labels"`
	Annotations          map[string]types.String               `tfsdk:"annotations"`
	DisplayName          types.String                          `tfsdk:"display_name"`
	Description          types.String                          `tfsdk:"description"`
	RetentionPolicyRules []branchImageRetentionPolicyRuleModel `tfsdk:"retention_policy_rules"`
}

func newBranchModel(obj *containerv1.Branch) branchModel {
	return branchModel{
		ID:                   types.StringValue(obj.Name),
		Name:                 types.StringValue(obj.Name),
		Labels:               conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations:          conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		DisplayName:          types.StringValue(obj.Spec.DisplayName),
		Description:          conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		RetentionPolicyRules: conv.EmptyIfNil(conv.ForEachSliceItem(obj.Spec.RetentionPolicyRules, newBranchImageRetentionPolicyRuleModel)),
	}
}

func (m branchModel) ToObject() *containerv1.Branch {
	return &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: containerv1.BranchSpec{
			DisplayName: m.DisplayName.ValueString(),
			Description: m.Description.ValueString(),
			RetentionPolicyRules: conv.ForEachSliceItem(m.RetentionPolicyRules,
				func(item branchImageRetentionPolicyRuleModel) containerv1.BranchImageRetentionPolicyRule {
					return item.ToObject()
				},
			),
		},
	}
}

type branchImageRetentionPolicyRuleModel struct {
	Name       types.String `tfsdk:"name"`
	ImageRegex types.String `tfsdk:"image_regex"`
	TagRegex   types.String `tfsdk:"tag_regex"`
	KeepCount  types.Int64  `tfsdk:"keep_count"`
	KeepDays   types.Int64  `tfsdk:"keep_days"`
}

func newBranchImageRetentionPolicyRuleModel(rule containerv1.BranchImageRetentionPolicyRule) branchImageRetentionPolicyRuleModel {
	return branchImageRetentionPolicyRuleModel{
		Name:       types.StringValue(rule.Name),
		ImageRegex: types.StringValue(rule.ImageRegex),
		TagRegex:   types.StringValue(rule.TagRegex),
		KeepCount:  types.Int64Value(int64(rule.KeepCount)),
		KeepDays:   types.Int64Value(int64(rule.KeepDays)),
	}
}

func (m branchImageRetentionPolicyRuleModel) ToObject() containerv1.BranchImageRetentionPolicyRule {
	return containerv1.BranchImageRetentionPolicyRule{
		Name:       m.Name.ValueString(),
		ImageRegex: m.ImageRegex.ValueString(),
		TagRegex:   m.TagRegex.ValueString(),
		KeepCount:  int(m.KeepCount.ValueInt64()),
		KeepDays:   int(m.KeepDays.ValueInt64()),
	}
}
