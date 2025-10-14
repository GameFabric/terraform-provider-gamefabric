package container

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// branchModel is the model for a branch resource.
type branchModel struct {
	ID                   types.String                          `tfsdk:"id"`
	Name                 types.String                          `tfsdk:"name"`
	DisplayName          types.String                          `tfsdk:"display_name"`
	Description          types.String                          `tfsdk:"description"`
	RetentionPolicyRules []branchImageRetentionPolicyRuleModel `tfsdk:"retention_policy_rules"`
}

func newBranchModel(obj *containerv1.Branch) branchModel {
	return branchModel{
		ID:          types.StringValue(obj.Name),
		Name:        types.StringValue(obj.Name),
		DisplayName: types.StringValue(obj.Spec.DisplayName),
		Description: conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		RetentionPolicyRules: conv.ForEachSliceItem(obj.Spec.RetentionPolicyRules, func(item containerv1.BranchImageRetentionPolicyRule) branchImageRetentionPolicyRuleModel {
			return branchImageRetentionPolicyRuleModel{
				Name:       types.StringValue(item.Name),
				ImageRegex: types.StringValue(item.ImageRegex),
				TagRegex:   types.StringValue(item.TagRegex),
				KeepCount:  types.Int64Value(int64(item.KeepCount)),
				KeepDays:   types.Int64Value(int64(item.KeepDays)),
			}
		}),
	}
}

// branchImageRetentionPolicyRuleModel is the model for an image retention policy rule.
type branchImageRetentionPolicyRuleModel struct {
	Name       types.String `tfsdk:"name"`
	ImageRegex types.String `tfsdk:"image_regex"`
	TagRegex   types.String `tfsdk:"tag_regex"`
	KeepCount  types.Int64  `tfsdk:"keep_count"`
	KeepDays   types.Int64  `tfsdk:"keep_days"`
}

func (m branchModel) ToObject() *containerv1.Branch {
	return &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name.ValueString(),
		},
		Spec: containerv1.BranchSpec{
			DisplayName:          m.DisplayName.ValueString(),
			Description:          m.Description.ValueString(),
			RetentionPolicyRules: conv.ForEachSliceItem(m.RetentionPolicyRules, retentionFromModel),
		},
	}
}
