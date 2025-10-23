package mps

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// HealthCheckAttributes returns the schema attributes for Agones health checks.
func HealthCheckAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"disabled": schema.BoolAttribute{
			Description:         "Disabled indicates whether Agones health checks are disabled.",
			MarkdownDescription: "Disabled indicates whether Agones health checks are disabled.",
			Optional:            true,
		},
		"period_seconds": schema.Int32Attribute{
			Description:         "PeriodSeconds is the number of seconds between checks.",
			MarkdownDescription: "PeriodSeconds is the number of seconds between checks.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
		"failure_threshold": schema.Int32Attribute{
			Description:         "FailureThreshold is the number of consecutive failures before the game server is marked unhealthy.",
			MarkdownDescription: "FailureThreshold is the number of consecutive failures before the game server is marked unhealthy.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
		"initial_delay_seconds": schema.Int32Attribute{
			Description:         "InitialDelaySeconds is the number of seconds to wait before performing the first check.",
			MarkdownDescription: "InitialDelaySeconds is the number of seconds to wait before performing the first check.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.AtLeast(0),
			},
		},
	}
}
