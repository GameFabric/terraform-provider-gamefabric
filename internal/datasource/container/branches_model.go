package container

import (
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type branchesModel struct {
	Name        types.String            `tfsdk:"name"`
	DisplayName types.String            `tfsdk:"display_name"`
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Branches    []branchModel           `tfsdk:"branches"`
}

func newBranchesModel(objs []containerv1.Branch) branchesModel {
	return branchesModel{
		Branches: conv.ForEachSliceItem(objs, func(item containerv1.Branch) branchModel {
			return newBranchModel(&item)
		}),
	}
}
