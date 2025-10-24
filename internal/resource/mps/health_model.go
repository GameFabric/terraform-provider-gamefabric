package mps

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HealthChecksModel is the terraform model for Agones health checks.
type HealthChecksModel struct {
	Disabled            types.Bool  `tfsdk:"disabled"`
	InitialDelaySeconds types.Int32 `tfsdk:"initial_delay_seconds"`
	PeriodSeconds       types.Int32 `tfsdk:"period_seconds"`
	FailureThreshold    types.Int32 `tfsdk:"failure_threshold"`
}

// NewHealthChecks converts the backend resource into the terraform model.
func NewHealthChecks(obj agonesv1.Health) *HealthChecksModel {
	return &HealthChecksModel{
		Disabled:            types.BoolValue(obj.Disabled),
		InitialDelaySeconds: types.Int32Value(obj.InitialDelaySeconds),
		PeriodSeconds:       types.Int32Value(obj.PeriodSeconds),
		FailureThreshold:    types.Int32Value(obj.FailureThreshold),
	}
}

// Default returns the default values for HealthChecksModel.
func (m HealthChecksModel) Default() defaults.Object {
	return objectdefault.StaticValue(types.ObjectValueMust(map[string]attr.Type{
		"disabled":              types.BoolType,
		"initial_delay_seconds": types.Int32Type,
		"period_seconds":        types.Int32Type,
		"failure_threshold":     types.Int32Type,
	}, map[string]attr.Value{
		"disabled":              types.BoolValue(false),
		"initial_delay_seconds": types.Int32Value(0),
		"period_seconds":        types.Int32Value(0),
		"failure_threshold":     types.Int32Value(0),
	}))
}

// ToHealthChecks converts the terraform model into the backend resource.
func ToHealthChecks(hc *HealthChecksModel) agonesv1.Health {
	if hc == nil {
		return agonesv1.Health{}
	}
	return agonesv1.Health{
		Disabled:            hc.Disabled.ValueBool(),
		InitialDelaySeconds: hc.InitialDelaySeconds.ValueInt32(),
		PeriodSeconds:       hc.PeriodSeconds.ValueInt32(),
		FailureThreshold:    hc.FailureThreshold.ValueInt32(),
	}
}
