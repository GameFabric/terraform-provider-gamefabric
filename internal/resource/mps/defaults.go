package mps

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultMapOf returns a default value for a map with the given element type.
func DefaultMapOf(typ attr.Type) defaults.Map {
	return mapdefault.StaticValue(types.MapValueMust(typ, nil))
}

// DefaultListOf returns a default value for a list with the given element type.
func DefaultListOf(typ attr.Type) defaults.List {
	return listdefault.StaticValue(types.ListValueMust(typ, nil))
}
