package conv

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Quantity converts a terraform string to a quantity.
func Quantity(val types.String) *resource.Quantity {
	if !IsKnown(val) {
		return nil
	}
	q, _ := resource.ParseQuantity(val.ValueString()) // Expected to be validated beforehand.
	return &q
}
