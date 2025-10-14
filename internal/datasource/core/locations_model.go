package core

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type locationsModel struct {
	Type      types.String   `tfsdk:"type"`
	City      types.String   `tfsdk:"city"`
	Country   types.String   `tfsdk:"country"`
	Continent types.String   `tfsdk:"continent"`
	NameRegex types.String   `tfsdk:"name_regex"`
	Locations []types.String `tfsdk:"locations"`
}

func newLocationsModel(locations []string) locationsModel {
	return locationsModel{
		Locations: conv.ForEachSliceItem(locations, func(item string) types.String { return types.StringValue(item) }),
	}
}
