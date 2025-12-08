package core

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// EnvVarAttributes returns the schema attributes for an environment variable.
func EnvVarAttributes(val validators.GameFabricValidator, pathPrefix string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description:         "Name is the name of the environment variable.",
			MarkdownDescription: "Name is the name of the environment variable.",
			Required:            true,
			Validators: []validator.String{
				validators.GFFieldString(val, pathPrefix+".name"),
			},
		},
		"value": schema.StringAttribute{
			Description:         "Value is the value of the environment variable.",
			MarkdownDescription: "Value is the value of the environment variable.",
			Optional:            true,
			Validators: []validator.String{
				validators.GFFieldString(val, pathPrefix+".value"),
				stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_from")),
			},
		},
		"value_from": schema.SingleNestedAttribute{
			Description:         "ValueFrom is the source for the environment variable&#39;s value.",
			MarkdownDescription: "ValueFrom is the source for the environment variable&#39;s value.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"field_path": schema.StringAttribute{
					Description:         "FieldPath selects the field of the pod. Supports `metadata.name`, `metadata.namespace`, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, `metadata.armadaName`, `metadata.regionName`, `metadata.regionTypeName`, `metadata.siteName`, `metadata.imageBranch`, `metadata.imageName`, `metadata.imageTag`, `spec.nodeName`, `spec.serviceAccountName`, `status.hostIP`, `status.podIP`, `status.podIPs`.",
					MarkdownDescription: "FieldPath selects the field of the pod. Supports `metadata.name`, `metadata.namespace`, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, `metadata.armadaName`, `metadata.regionName`, `metadata.regionTypeName`, `metadata.siteName`, `metadata.imageBranch`, `metadata.imageName`, `metadata.imageTag`, `spec.nodeName`, `spec.serviceAccountName`, `status.hostIP`, `status.podIP`, `status.podIPs`.",
					Optional:            true,
					Validators: []validator.String{
						validators.GFFieldString(val, pathPrefix+".valueFrom.fieldRef.fieldPath"),
						stringvalidator.ExactlyOneOf(
							path.MatchRelative().AtParent().AtName("config_file"),
							path.MatchRelative().AtParent().AtName("secret"),
						),
					},
				},
				"config_file": schema.StringAttribute{
					Description:         "ConfigFile select the configuration file.",
					MarkdownDescription: "ConfigFile select the configuration file.",
					Optional:            true,
					Validators: []validator.String{
						validators.GFFieldString(val, pathPrefix+".valueFrom.configFileKeyRef.name"),
						stringvalidator.ExactlyOneOf(
							path.MatchRelative().AtParent().AtName("field_path"),
							path.MatchRelative().AtParent().AtName("secret"),
						),
					},
				},
				"secret": schema.SingleNestedAttribute{
					Description:         "Secret selects the secret.",
					MarkdownDescription: "Secret selects the secret.",
					Optional:            true,
					Validators: []validator.Object{
						validators.GFFieldObject(val, pathPrefix+".valueFrom.secretKeyRef.name"),
						objectvalidator.ExactlyOneOf(
							path.MatchRelative().AtParent().AtName("config_file"),
							path.MatchRelative().AtParent().AtName("field_path"),
						),
					},
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description:         "Key of the secret.",
							MarkdownDescription: "Key of the secret.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							Description:         "Name of the secret.",
							MarkdownDescription: "Name of the secret.",
							Required:            true,
						},
					},
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value")),
			},
		},
	}
}
