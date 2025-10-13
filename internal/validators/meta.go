package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	maxNameLength        = 63
	maxEnvironmentLength = 4
)

var (
	nameRegexp        = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	environmentRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

// NameValidator is a custom validator that checks if a string is a valid name.
type NameValidator struct{}

// Description provides a description of the validator.
func (n NameValidator) Description(context.Context) string {
	return fmt.Sprintf("Validates that the value is a valid name with a maximum length of %d characters.", maxNameLength)
}

// MarkdownDescription provides a markdown description of the validator.
func (n NameValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("Validates that the value is a valid name with a maximum length of %d characters.", maxNameLength)
}

// ValidateString checks that the provided string is a valid name.
func (n NameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	v := req.ConfigValue.ValueString()

	if len(v) > maxNameLength {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid name length",
			fmt.Sprintf("%q must be no more than %d characters", v, maxNameLength),
		))
	}
	if v != "" && !nameRegexp.MatchString(v) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid name",
			v+` is not a valid name`,
		))
	}
}

// EnvironmentValidator is a custom validator that checks if a string is a valid environment name.
type EnvironmentValidator struct{}

// Description provides a description of the validator.
func (n EnvironmentValidator) Description(context.Context) string {
	return fmt.Sprintf("Validates that the value is a valid name with a maximum length of %d characters.", maxNameLength)
}

// MarkdownDescription provides a markdown description of the validator.
func (n EnvironmentValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("Validates that the value is a valid name with a maximum length of %d characters.", maxNameLength)
}

// ValidateString checks that the provided string is a valid name.
func (n EnvironmentValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	v := req.ConfigValue.ValueString()

	if len(v) > maxEnvironmentLength {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid environment name length",
			fmt.Sprintf("%q must be no more than %d characters", v, maxEnvironmentLength),
		))
	}
	if v != "" && !environmentRegexp.MatchString(v) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid environment name",
			v+` is not a valid environment name`,
		))
	}
}
