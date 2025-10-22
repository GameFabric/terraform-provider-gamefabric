package container

import (
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageModel struct {
	Name   types.String `tfsdk:"name"`
	Branch types.String `tfsdk:"branch"`
	Image  types.String `tfsdk:"image"`
	Tag    types.String `tfsdk:"tag"`
}

func newImageModel(obj containerv1.Image) imageModel {
	return imageModel{
		Name:   types.StringValue(obj.Name),
		Branch: types.StringValue(obj.Branch),
		Image:  types.StringValue(obj.Spec.Image),
		Tag:    types.StringValue(obj.Spec.Tag),
	}
}
