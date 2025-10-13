package container

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BranchModel is the model for a branch resource.
type BranchModel struct {
	Name                 types.String                          `tfsdk:"name"`
	DisplayName          types.String                          `tfsdk:"display_name"`
	Description          types.String                          `tfsdk:"description"`
	RetentionPolicyRules []BranchImageRetentionPolicyRuleModel `tfsdk:"retention_policy_rules"`
}

// BranchImageRetentionPolicyRuleModel is the model for an image retention policy rule.
type BranchImageRetentionPolicyRuleModel struct {
	Name       types.String `tfsdk:"name"`
	ImageRegex types.String `tfsdk:"image_regex"`
	TagRegex   types.String `tfsdk:"tag_regex"`
	KeepCount  types.Int64  `tfsdk:"keep_count"`
	KeepDays   types.Int64  `tfsdk:"keep_days"`
}

// BranchAttributes returns the schema attributes for a branch resource.
func BranchAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description:         "The unique object name within its scope.",
			MarkdownDescription: "The unique object name within its scope.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				validators.NameValidator{},
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"display_name": schema.StringAttribute{
			Description:         "DisplayName is the display name of the branch.",
			MarkdownDescription: "DisplayName is the display name of the branch.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			Description:         "Description is the optional description of the branch.",
			MarkdownDescription: "Description is the optional description of the branch.",
			Optional:            true,
		},
		"retention_policy_rules": schema.ListNestedAttribute{
			Description:         "RetentionPolicyRules are the rules that define how images are retained.",
			MarkdownDescription: "RetentionPolicyRules are the rules that define how images are retained.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description:         "Name is the name of the image retention policy.",
						MarkdownDescription: "Name is the name of the image retention policy.",
						Required:            true,
					},
					"image_regex": schema.StringAttribute{
						Description:         "ImageRegex is the optional regex selector for images that this policy applies to.",
						MarkdownDescription: "ImageRegex is the optional regex selector for images that this policy applies to.",
						Optional:            true,
					},
					"tag_regex": schema.StringAttribute{
						Description:         "TagRegex is the optional regex selector for tags that this policy applies to.",
						MarkdownDescription: "TagRegex is the optional regex selector for tags that this policy applies to.",
						Optional:            true,
					},
					"keep_count": schema.Int64Attribute{
						Description:         "KeepCount is the minimum number of tags to keep per image.",
						MarkdownDescription: "KeepCount is the minimum number of tags to keep per image.",
						Optional:            true,
					},
					"keep_days": schema.Int64Attribute{
						Description:         "KeepDays is the minimum number of days an image tag must be kept for.",
						MarkdownDescription: "KeepDays is the minimum number of days an image tag must be kept for.",
						Optional:            true,
					},
				},
			},
		},
	}
}
