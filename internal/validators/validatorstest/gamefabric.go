package validatorstest

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/gamefabric/gf-apicore/runtime"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// CollectPathExpressions collects all path expressions from the given schema.
//
// Path expressions for wrapped validators can not be exposed.
func CollectPathExpressions(s schema.Schema) []string {
	var res []string
	for _, attr := range s.Attributes {
		res = append(res, collectPathExprs(attr, "")...)
	}
	slices.Sort(res)
	return res
}

// CollectJSONPaths collects all JSON paths from the given runtime object.
func CollectJSONPaths(obj runtime.Object) []string {
	res := collectJSONPaths(reflect.TypeOf(obj), "")
	slices.Sort(res)
	return res
}

func collectPathExprs(attr schema.Attribute, prefix string) []string { //nolint:cyclop // Can not split the switch.
	var res []string
	switch typ := attr.(type) {
	case schema.BoolAttribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
	case schema.Int32Attribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
	case schema.Int64Attribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
	case schema.StringAttribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
	case schema.MapAttribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
	case schema.ListNestedAttribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
		for _, val := range typ.NestedObject.Validators {
			res = append(res, pathExprs(val)...)
		}
		for key, subAttr := range typ.NestedObject.Attributes {
			res = append(res, collectPathExprs(subAttr, strings.TrimLeft(prefix+".", ".")+key)...)
		}
	case schema.ListAttribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
	case schema.SingleNestedAttribute:
		for _, val := range typ.Validators {
			res = append(res, pathExprs(val)...)
		}
		for key, subAttr := range typ.Attributes {
			res = append(res, collectPathExprs(subAttr, strings.TrimLeft(prefix+".", ".")+key)...)
		}
	default:
		panic(fmt.Sprintf("unknown attribute type: %T", attr))
	}
	return res
}

func pathExprs(val validator.Describer) []string {
	gfVal, ok := val.(interface{ PathExpr() string })
	if !ok {
		return nil
	}
	return []string{gfVal.PathExpr()}
}

func collectJSONPaths(t reflect.Type, prefix string) []string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var res []string
	if prefix != "" {
		res = append(res, prefix)
	}

	// Special handling.
	if strings.HasSuffix(prefix, ".resources") {
		res = append(res,
			prefix+".limits.cpu",
			prefix+".limits.memory",
			prefix+".requests.cpu",
			prefix+".requests.memory",
		)
		return res
	}

	switch t.Kind() {
	case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint16:
	case reflect.Struct:
		for i := range t.NumField() {
			field := t.Field(i)
			json := field.Tag.Get("json")
			if json == "-" || field.Tag == "" {
				continue
			}
			json, _, _ = strings.Cut(json, ",") // Cut omitempty or similar.

			if json == "" {
				res = append(res, collectJSONPaths(field.Type, prefix)...) // Embed.
				continue
			}

			res = append(res, collectJSONPaths(field.Type, joinPath(prefix, json))...)
		}
	case reflect.Map:
		res = append(res, collectJSONPaths(t.Elem(), prefix+"[?]")...)
	case reflect.Slice:
		res = append(res, collectJSONPaths(t.Elem(), prefix+"[?]")...)
	default:
		panic("unhandled kind: " + t.Kind().String())
	}

	return res
}

func joinPath(a, b string) string {
	if a == "" {
		return b
	}
	return a + "." + b
}
