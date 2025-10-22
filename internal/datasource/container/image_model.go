package container

import (
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageModel struct {
	Branch      types.String `tfsdk:"branch"`
	Image       types.String `tfsdk:"image"`
	Tag         types.String `tfsdk:"tag"`
	ImageTarget *imageTarget `tfsdk:"image_target"`
}

func newImageModel(obj containerv1.Image) imageModel {
	return imageModel{
		ImageTarget: &imageTarget{
			Name:   types.StringValue(obj.Name),
			Branch: types.StringValue(obj.Branch),
		},
	}
}

type imageTarget struct {
	Name   types.String `tfsdk:"name"`
	Branch types.String `tfsdk:"branch"`
}
