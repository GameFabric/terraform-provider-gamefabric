package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinReplicasValidator validates the relationship between `min_replicas`,
// `buffer_size` and `dynamic_buffer` attributes within a replicas list entry.
//
// Rules:
//   - A value of 0 is always allowed and instructs the API to use the buffer
//     value as the effective minimum.
//   - When the value is non-zero it must be >= buffer_size.
//   - When dynamic_buffer is configured the validator emits warnings to
//     guide the user toward the most useful configuration:
//   - min_replicas == buffer_size: suggest setting min_replicas to 0 to
//     give the dynamic buffer full control.
//   - min_replicas >  buffer_size: explain that dynamic buffer cannot
//     grow the buffer beyond min_replicas.
//   - min_replicas == 0: no warning is emitted.
type MinReplicasValidator struct{}

// Description describes the validation in plain text formatting.
func (v MinReplicasValidator) Description(_ context.Context) string {
	return "value must be 0 or greater than or equal to buffer_size"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v MinReplicasValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateInt32 performs the validation.
//
//nolint:cyclop // Belongs together.
func (v MinReplicasValidator) ValidateInt32(ctx context.Context, request validator.Int32Request, response *validator.Int32Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	minReplicas := request.ConfigValue.ValueInt32()

	parent := request.Path.ParentPath()

	// Read sibling buffer_size.
	var bufferAttr attr.Value
	diags := request.Config.GetAttribute(ctx, parent.AtName("buffer_size"), &bufferAttr)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	if bufferAttr == nil || bufferAttr.IsUnknown() {
		return
	}

	var bufferSize int32
	if !bufferAttr.IsNull() {
		var b types.Int32
		d := tfsdk.ValueAs(ctx, bufferAttr, &b)
		response.Diagnostics.Append(d...)
		if d.HasError() {
			return
		}
		bufferSize = b.ValueInt32()
	}

	// Hard validation: when min_replicas is non-zero it must be >= buffer_size.
	if minReplicas != 0 && minReplicas < bufferSize {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Attribute Value",
			fmt.Sprintf("Attribute %s value must be 0 or at least equal to buffer_size (%d), got: %d. "+
				"A value of 0 instructs GameFabric to use buffer_size as the effective minimum.",
				request.Path, bufferSize, minReplicas),
		)
		return
	}

	// Determine whether dynamic_buffer is configured.
	var dynBufAttr attr.Value
	diags = request.Config.GetAttribute(ctx, parent.AtName("dynamic_buffer"), &dynBufAttr)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	if dynBufAttr == nil || dynBufAttr.IsUnknown() || dynBufAttr.IsNull() {
		return
	}

	// Dynamic buffer is set: emit warnings to help the user.
	switch {
	case minReplicas == 0:
		// No warning: this gives full control to dynamic buffer.
	case minReplicas == bufferSize:
		response.Diagnostics.AddAttributeWarning(
			request.Path,
			"Consider setting min_replicas to 0",
			"min_replicas is equal to buffer_size while dynamic_buffer is enabled. "+
				"In this configuration min_replicas dictates the maximum size dynamic_buffer can apply as buffer. "+
				"Setting min_replicas to 0 gives full control to dynamic_buffer and is the recommended value when "+
				"you only care about buffer.",
		)
	case minReplicas > bufferSize:
		response.Diagnostics.AddAttributeWarning(
			request.Path,
			"min_replicas caps the dynamic buffer size",
			fmt.Sprintf("dynamic_buffer is enabled but min_replicas (%d) is greater than buffer_size (%d). "+
				"Note that dynamic_buffer cannot make the effective buffer larger than min_replicas. "+
				"Set min_replicas to 0 to give dynamic_buffer full control.",
				minReplicas, bufferSize),
		)
	}
}
