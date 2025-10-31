package rbac

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleModel struct {
	ID     types.String            `tfsdk:"id"`
	Name   types.String            `tfsdk:"name"`
	Rules  []ruleModel             `tfsdk:"rules"`
	Labels map[string]types.String `tfsdk:"labels"`
}

type ruleModel struct {
	Verbs         []types.String `tfsdk:"verbs"`
	APIGroups     []types.String `tfsdk:"api_groups"`
	Environments  []types.String `tfsdk:"environments"`
	Resources     []types.String `tfsdk:"resources"`
	Scopes        []types.String `tfsdk:"scopes"`
	ResourceNames []types.String `tfsdk:"resource_names"`
}

func newRoleModel(obj *rbacv1.Role) roleModel {
	return roleModel{
		ID:     types.StringValue(obj.Name),
		Name:   types.StringValue(obj.Name),
		Labels: conv.ForEachMapItem(obj.Labels, types.StringValue),
		Rules: conv.ForEachSliceItem(obj.Rules, func(rule rbacv1.Rule) ruleModel {
			return ruleModel{
				Verbs:         conv.ForEachSliceItem(rule.Verbs, types.StringValue),
				APIGroups:     conv.ForEachSliceItem(rule.APIGroups, types.StringValue),
				Environments:  conv.ForEachSliceItem(rule.Environments, types.StringValue),
				Resources:     conv.ForEachSliceItem(rule.Resources, types.StringValue),
				Scopes:        conv.ForEachSliceItem(rule.Scopes, types.StringValue),
				ResourceNames: conv.ForEachSliceItem(rule.ResourceNames, types.StringValue),
			}
		}),
	}
}

func (m roleModel) ToObject() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   m.Name.ValueString(),
			Labels: conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
		},
		Rules: conv.ForEachSliceItem(m.Rules, func(rule ruleModel) rbacv1.Rule {
			return rbacv1.Rule{
				Verbs:         conv.ForEachSliceItem(rule.Verbs, func(item types.String) string { return item.ValueString() }),
				APIGroups:     conv.ForEachSliceItem(rule.APIGroups, func(item types.String) string { return item.ValueString() }),
				Environments:  conv.ForEachSliceItem(rule.Environments, func(item types.String) string { return item.ValueString() }),
				Resources:     conv.ForEachSliceItem(rule.Resources, func(item types.String) string { return item.ValueString() }),
				Scopes:        conv.ForEachSliceItem(rule.Scopes, func(item types.String) string { return item.ValueString() }),
				ResourceNames: conv.ForEachSliceItem(rule.ResourceNames, func(item types.String) string { return item.ValueString() }),
			}
		}),
	}
}
