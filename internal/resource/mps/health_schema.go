package mps

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// HealthCheckAttributes returns the schema attributes for Agones health checks.
func HealthCheckAttributes(val validators.GameFabricValidator, pathPrefix string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"disabled": schema.BoolAttribute{
			Description:         "Disabled indicates whether Agones health checks are disabled.",
			MarkdownDescription: "Disabled indicates whether Agones health checks are disabled.",
			Optional:            true,
			Validators: []validator.Bool{
				validators.GFFieldBool(val, pathPrefix+".disabled"),
			},
		},
		"period_seconds": schema.Int32Attribute{
			Description:         "PeriodSeconds is the number of seconds between checks.",
			MarkdownDescription: "PeriodSeconds is the number of seconds between checks.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.AtLeast(0),
				validators.GFFieldInt32(val, pathPrefix+".periodSeconds"),
			},
		},
		"failure_threshold": schema.Int32Attribute{
			Description:         "FailureThreshold is the number of consecutive failures before the game server is marked unhealthy and terminated.",
			MarkdownDescription: "FailureThreshold is the number of consecutive failures before the game server is marked unhealthy and terminated.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.AtLeast(0),
				validators.GFFieldInt32(val, pathPrefix+".failureThreshold"),
			},
		},
		"initial_delay_seconds": schema.Int32Attribute{
			Description:         "InitialDelaySeconds is the number of seconds to wait before performing the first check.",
			MarkdownDescription: "InitialDelaySeconds is the number of seconds to wait before performing the first check.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.AtLeast(0),
				validators.GFFieldInt32(val, pathPrefix+".initialDelaySeconds"),
			},
		},
	}
}
