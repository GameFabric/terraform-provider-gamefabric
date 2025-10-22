package conv

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ForEachSliceItem applies a function to each item in a slice and returns a new slice with the results.
func ForEachSliceItem[T, K any](s []T, fn func(item T) K) []K {
	if len(s) == 0 {
		return nil
	}
	result := make([]K, len(s))
	for i, item := range s {
		result[i] = fn(item)
	}
	return result
}

// ForEachMapItem applies a function to each item in a map and returns a new map with the results.
func ForEachMapItem[T, K any](m map[string]T, fn func(item T) K) map[string]K {
	if len(m) == 0 {
		return nil
	}
	result := make(map[string]K, len(m))
	for key, item := range m {
		result[key] = fn(item)
	}
	return result
}

// IsKnown checks if the value is neither null nor unknown.
func IsKnown(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// OptionalFunc checks if the value is empty and returns a default value, from defFn, if it is,
// otherwise it applies valFn to the value.
func OptionalFunc[T comparable, K attr.Value](v T, valFn func(T) K, defFn func() K) K {
	var emptyT T
	if v == emptyT {
		return defFn()
	}
	return valFn(v)
}

// EmptyIfNil returns an empty slice if the input slice is nil, otherwise it returns the input slice.
func EmptyIfNil[T any](v []T) []T {
	if v == nil {
		return make([]T, 0)
	}
	return v
}

// FromIntOrString converts an IntOrString to a Terraform String type.
func FromIntOrString(val *intstr.IntOrString) types.String {
	if val == nil {
		return types.StringNull()
	}
	if val.Type == intstr.Int {
		return types.StringValue(strconv.Itoa(int(val.IntVal)))
	}
	return types.StringValue(val.StrVal)
}
