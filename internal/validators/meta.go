package validators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gamefabric/gf-apicore/api/validation"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	maxNameLength             = 63
	maxEnvironmentLength      = 4
	maxLabelValueLength       = 63
	totalAnnotationSizeLimitB = 256 * (1 << 10) // 256 kB
)

var (
	nameRegexp        = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	environmentRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	labelValueRegexp  = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-_.]*)?[A-Za-z0-9]$`)
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
	if !conv.IsKnown(req.ConfigValue) {
		// The framework still calls the validator when the value is optional and not set.
		// This means we have to skip null values as well.
		return
	}

	value := req.ConfigValue.ValueString()
	switch {
	case len(value) > maxNameLength:
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid name length",
			fmt.Sprintf("%s must be no more than %d characters", value, maxNameLength),
		))
	case value != "" && !nameRegexp.MatchString(value):
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid name",
			value+` is not a valid name`,
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

// LabelsValidator is a custom validator that checks if a map's keys and values are valid.
type LabelsValidator struct{}

// Description provides a description of the validator.
func (n LabelsValidator) Description(context.Context) string {
	return "Validates that the label keys are fully qualified and the values are valid."
}

// MarkdownDescription provides a markdown description of the validator.
func (n LabelsValidator) MarkdownDescription(context.Context) string {
	return "Validates that the label keys are fully qualified and the values are valid."
}

// ValidateMap checks that the provided map's keys and values are valid.
func (n LabelsValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for key, val := range req.ConfigValue.Elements() {
		if keyErrs := validation.IsQualifiedName(key); len(keyErrs) > 0 {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Invalid label key",
				fmt.Sprintf("Label key %q is not valid: %s", key, strings.Join(keyErrs, "; ")),
			))
		}

		strVal, ok := val.(basetypes.StringValue)
		if !ok {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Invalid label value type",
				fmt.Sprintf("Label value for key %q is not a string", key),
			))
			continue
		}
		if strVal.IsNull() || strVal.IsUnknown() {
			continue
		}

		if len(strVal.String()) > maxLabelValueLength {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Invalid label value length",
				fmt.Sprintf("Label value for key %q must be no more than %d characters", key, maxLabelValueLength),
			))
		}
		if !labelValueRegexp.MatchString(strVal.ValueString()) {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Invalid label value",
				fmt.Sprintf("Label value for key %q is not valid", key),
			))
		}
	}
}

// AnnotationsValidator is a custom validator that checks if a map's keys and values are valid.
type AnnotationsValidator struct{}

// Description provides a description of the validator.
func (n AnnotationsValidator) Description(context.Context) string {
	return "Validates that the annotation keys are fully qualified and the values are valid."
}

// MarkdownDescription provides a markdown description of the validator.
func (n AnnotationsValidator) MarkdownDescription(context.Context) string {
	return "Validates that the annotation keys are fully qualified and the values are valid."
}

// ValidateMap checks that the provided map's keys and values are valid.
func (n AnnotationsValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var size int
	for key, val := range req.ConfigValue.Elements() {
		if keyErrs := validation.IsQualifiedName(key); len(keyErrs) > 0 {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Invalid annotation key",
				fmt.Sprintf("Annotation key %q is not valid: %s", key, strings.Join(keyErrs, "; ")),
			))
		}

		strVal, ok := val.(basetypes.StringValue)
		if !ok {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Invalid annotation value type",
				fmt.Sprintf("Annotation value for key %q is not a string", key),
			))
			continue
		}
		if strVal.IsNull() || strVal.IsUnknown() {
			continue
		}

		size += len(key) + len(strVal.ValueString())
	}
	if size > totalAnnotationSizeLimitB {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid annotations size",
			fmt.Sprintf("Total size of annotations must be no more than %d bytes", totalAnnotationSizeLimitB),
		))
	}
}
