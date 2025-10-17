package validators

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// CIDRValidator validates that a string is a valid CIDR.
type CIDRValidator struct{}

// Description provides a description of the validator.
func (v *CIDRValidator) Description(_ context.Context) string {
	return "Validates that the value is a valid Classless Inter-Domain Routing (CIDR)"
}

// MarkdownDescription provides a markdown description of the validator.
func (v *CIDRValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateList checks that each string in the list is a valid CIDR.
func (v *CIDRValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for _, elem := range req.ConfigValue.Elements() {
		if elem.IsUnknown() || elem.IsNull() {
			continue
		}

		val, ok := elem.(basetypes.StringValue)
		if !ok {
			resp.Diagnostics.AddError(
				"Invalid Type",
				"Expected element to be of type String.",
			)
			continue
		}

		if _, _, err := net.ParseCIDR(val.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Invalid CIDR",
				err.Error(),
			)
		}
	}
}
