package mps

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
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
	if obj == (agonesv1.Health{}) {
		return nil
	}

	return &HealthChecksModel{
		Disabled:            types.BoolValue(obj.Disabled),
		InitialDelaySeconds: conv.OptionalFunc(obj.InitialDelaySeconds, types.Int32Value, types.Int32Null),
		PeriodSeconds:       conv.OptionalFunc(obj.PeriodSeconds, types.Int32Value, types.Int32Null),
		FailureThreshold:    conv.OptionalFunc(obj.FailureThreshold, types.Int32Value, types.Int32Null),
	}
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
