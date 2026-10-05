package mps

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// metricPolicyValidator enforces cross-field rules for ports with policy="Metric":
//   - protocol must be TCP
//   - container_port must be set
type metricPolicyValidator struct{}

func (v metricPolicyValidator) Description(_ context.Context) string {
	return "Metric ports require protocol=TCP and a container_port."
}

func (v metricPolicyValidator) MarkdownDescription(_ context.Context) string {
	return "Metric ports require `protocol = \"TCP\"` and a `container_port`."
}

func (v metricPolicyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueString() != MetricPortPolicy {
		return
	}

	parentPath := req.Path.ParentPath()

	// protocol must be TCP.
	var protocol types.String
	diags := req.Config.GetAttribute(ctx, parentPath.AtName("protocol"), &protocol)
	resp.Diagnostics.Append(diags...)
	if !diags.HasError() && !protocol.IsNull() && !protocol.IsUnknown() && protocol.ValueString() != "TCP" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid protocol for Metric port",
			"Metric ports must use protocol \"TCP\", got: "+protocol.ValueString(),
		)
	}
	if !diags.HasError() && protocol.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("protocol"),
			"Missing protocol for Metric port",
			"Metric ports require protocol = \"TCP\".",
		)
	}

	// container_port must be set.
	var containerPort attr.Value
	diags = req.Config.GetAttribute(ctx, parentPath.AtName("container_port"), &containerPort)
	resp.Diagnostics.Append(diags...)
	if !diags.HasError() && containerPort.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing container_port for Metric port",
			"Metric ports require container_port to be set.",
		)
	}
}

// pathOnlyForMetricValidator rejects a non-null path value when policy is not Metric.
type pathOnlyForMetricValidator struct{}

func (v pathOnlyForMetricValidator) Description(_ context.Context) string {
	return "path is only valid when policy is Metric."
}

func (v pathOnlyForMetricValidator) MarkdownDescription(_ context.Context) string {
	return "`path` is only valid when `policy` is `Metric`."
}

func (v pathOnlyForMetricValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var policy types.String
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("policy"), &policy)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	if !policy.IsNull() && !policy.IsUnknown() && policy.ValueString() != MetricPortPolicy {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid attribute for non-Metric port",
			"path may only be set when policy is \"Metric\".",
		)
	}
}

// noCommaValidator rejects strings containing a comma, which would break the
// comma-delimited g8c.io/metrics-endpoints annotation format.
type noCommaValidator struct{}

func (v noCommaValidator) Description(_ context.Context) string {
	return "Value must not contain a comma."
}

func (v noCommaValidator) MarkdownDescription(_ context.Context) string {
	return "Value must not contain a comma (`,`)."
}

func (v noCommaValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.Contains(req.ConfigValue.ValueString(), ",") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid character in path",
			"Metrics path must not contain a comma — it is used as the annotation delimiter.",
		)
	}
}
