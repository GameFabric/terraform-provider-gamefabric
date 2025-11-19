package core

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type locationsModel struct {
	Type      types.String `tfsdk:"type"`
	City      types.String `tfsdk:"city"`
	Country   types.String `tfsdk:"country"`
	Continent types.String `tfsdk:"continent"`
	NameRegex types.String `tfsdk:"name_regex"`

	// Accepted types are map[string]string or map[string][]string.
	// Dynamic types inside of collections (map[string]Dynamic) are not currently supported in terraform-plugin-framework.
	LabelFilter types.Dynamic `tfsdk:"label_filter"`

	Names []types.String `tfsdk:"names"`
}

func newLocationsModel(locations []string) locationsModel {
	return locationsModel{
		Names: conv.ForEachSliceItem(locations, func(item string) types.String { return types.StringValue(item) }),
	}
}
