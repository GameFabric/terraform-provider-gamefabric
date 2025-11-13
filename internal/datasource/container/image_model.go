package container

import (
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageModel struct {
	Branch   types.String `tfsdk:"branch"`
	Image    types.String `tfsdk:"image"`
	Tag      types.String `tfsdk:"tag"`
	ImageRef *imageRef    `tfsdk:"image_ref"`
}

func newImageModel(obj containerv1.Image) imageModel {
	return imageModel{
		ImageRef: &imageRef{
			Name:   types.StringValue(obj.Name),
			Branch: types.StringValue(obj.Branch),
		},
	}
}

type imageRef struct {
	Name   types.String `tfsdk:"name"`
	Branch types.String `tfsdk:"branch"`
}
