package container

import (
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type branchModel struct {
	Name                 types.String                          `tfsdk:"name"`
	DisplayName          types.String                          `tfsdk:"display_name"`
	RetentionPolicyRules []branchImageRetentionPolicyRuleModel `tfsdk:"retention_policy_rules"`
}

func newBranchModel(obj *containerv1.Branch) branchModel {
	return branchModel{
		Name:        types.StringValue(obj.Name),
		DisplayName: types.StringValue(obj.Spec.DisplayName),
		RetentionPolicyRules: conv.ForEachSliceItem(obj.Spec.RetentionPolicyRules,
			func(item containerv1.BranchImageRetentionPolicyRule) branchImageRetentionPolicyRuleModel {
				return branchImageRetentionPolicyRuleModel{
					Name:       types.StringValue(item.Name),
					ImageRegex: types.StringValue(item.ImageRegex),
					TagRegex:   types.StringValue(item.TagRegex),
					KeepCount:  types.Int64Value(int64(item.KeepCount)),
					KeepDays:   types.Int64Value(int64(item.KeepDays)),
				}
			},
		),
	}
}

type branchImageRetentionPolicyRuleModel struct {
	Name       types.String `tfsdk:"name"`
	ImageRegex types.String `tfsdk:"image_regex"`
	TagRegex   types.String `tfsdk:"tag_regex"`
	KeepCount  types.Int64  `tfsdk:"keep_count"`
	KeepDays   types.Int64  `tfsdk:"keep_days"`
}
