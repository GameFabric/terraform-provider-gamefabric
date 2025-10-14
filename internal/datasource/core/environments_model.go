package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type environmentsModel struct {
	LabelFilter  map[string]types.String `tfsdk:"label_filter"`
	Environments []environmentModel      `tfsdk:"environments"`
}

func newEnvironmentsModel(objs []corev1.Environment) environmentsModel {
	return environmentsModel{
		Environments: conv.ForEachSliceItem(objs, func(item corev1.Environment) environmentModel {
			return newEnvironmentModel(&item)
		}),
	}
}
