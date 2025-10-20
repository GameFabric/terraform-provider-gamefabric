package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"k8s.io/apimachinery/pkg/api/resource"
)

type QuantityValidator struct{}

func (v QuantityValidator) Description(_ context.Context) string {
	return "Validates that the attribute value is a valid resource identifier."
}

func (v QuantityValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v QuantityValidator) ValidateString(_ context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()
	q, err := resource.ParseQuantity(val)
	if err != nil {
		res.Diagnostics.AddError(
			"Invalid resource quantity",
			fmt.Sprintf("Resource quantity %q is invalid: %s", val, err.Error()),
		)
		return
	}

	if q.IsZero() || q.Sign() <= 0 {
		res.Diagnostics.AddError(
			"Invalid resource quantity",
			fmt.Sprintf("Resource quantity %q must be greater than zero", val),
		)
		return
	}

	// TODO we need to cross check requests < limits, probably not here but somewhere else.
}
