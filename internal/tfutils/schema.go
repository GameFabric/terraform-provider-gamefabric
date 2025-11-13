package tfutils

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WalkResourceSchema walks all attributes in the given resource schema, calling
// fn for each attribute with its path.
func WalkResourceSchema(schema rschema.Schema, fn func(rschema.Attribute, path.Path)) {
	walkResourceSchemaAtPath(schema.Attributes, path.Empty(), fn)
}

func walkResourceSchemaAtPath(attrs map[string]rschema.Attribute, p path.Path, fn func(rschema.Attribute, path.Path)) {
	for name, attr := range attrs {
		p := p.AtName(name)

		fn(attr, p)

		if nestedAttr, ok := attr.(rschema.NestedAttribute); ok {
			switch nestedAttr.GetNestingMode() {
			case 2: // ListNested
				p = p.AtListIndex(0)
			case 3: // SetNested
				p = p.AtSetValue(types.StringValue("*"))
			case 4: // MapNested
				p = p.AtMapKey("*")
			}

			attrObj, ok := nestedAttr.GetNestedObject().(rschema.NestedAttributeObject)
			if !ok {
				// Should not happen, but guard against panic.
				continue
			}
			walkResourceSchemaAtPath(attrObj.Attributes, p, fn)
		}
	}
}
