package validators

import (
	"context"
	"fmt"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"k8s.io/apimachinery/pkg/api/resource"
)

// QuantityValidator validates that a string is a valid resource quantity.
type QuantityValidator struct{}

// Description provides a description of the validator.
func (v QuantityValidator) Description(_ context.Context) string {
	return "Validates that the attribute value is a valid resource quantity."
}

// MarkdownDescription provides a markdown formatted description of the validator.
func (v QuantityValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString checks that the string is a valid resource quantity.
func (v QuantityValidator) ValidateString(_ context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if !conv.IsKnown(req.ConfigValue) {
		return
	}

	val := req.ConfigValue.ValueString()
	qty, err := resource.ParseQuantity(val)
	if err != nil {
		res.Diagnostics.AddError(
			"Invalid resource quantity",
			fmt.Sprintf("Resource quantity %q is invalid: %s", val, err.Error()),
		)
		return
	}
	if qty.Cmp(resource.MustParse(val)) != 0 {
		res.Diagnostics.AddError(
			"Invalid resource quantity format",
			fmt.Sprintf("Resource quantity %q is not in canonical form, should be %q", val, qty.String()),
		)
	}
}
