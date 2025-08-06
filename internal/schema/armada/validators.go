package armada

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"k8s.io/apimachinery/pkg/api/resource"
)

type quantityValidator struct{}

func (n quantityValidator) Description(context.Context) string {
	return "Validates that the value is a valid resource quantity."
}

func (n quantityValidator) MarkdownDescription(context.Context) string {
	return "Validates that the value is a valid resource quantity."
}

func (n quantityValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	_, err := resource.ParseQuantity(req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid quantity",
			err.Error(),
		))
	}
}
