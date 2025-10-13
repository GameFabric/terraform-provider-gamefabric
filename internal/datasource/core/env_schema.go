package core

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

// EnvVarAttributes returns the attributes for EnvVar.
func EnvVarAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description:         "Name is the name of the environment variable.",
			MarkdownDescription: "Name is the name of the environment variable.",
			Computed:            true,
		},
		"value": schema.StringAttribute{
			Description:         "Value is the value of the environment variable.",
			MarkdownDescription: "Value is the value of the environment variable.",
			Computed:            true,
		},
		"value_from": schema.SingleNestedAttribute{
			Description:         "ValueFrom is the source for the environment variable&#39;s value.",
			MarkdownDescription: "ValueFrom is the source for the environment variable&#39;s value.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"field_ref": schema.SingleNestedAttribute{
					Description:         "FieldRef selects the field of the pod. Supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`, metadata.armadaName, metadata.regionName, metadata.regionTypeName, metadata.siteName, metadata.imageBranch, metadata.imageName, metadata.imageTag, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.",
					MarkdownDescription: "FieldRef selects the field of the pod. Supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`, metadata.armadaName, metadata.regionName, metadata.regionTypeName, metadata.siteName, metadata.imageBranch, metadata.imageName, metadata.imageTag, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"api_version": schema.StringAttribute{
							Computed: true,
						},
						"field_path": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				"config_file_key_ref": schema.SingleNestedAttribute{
					Description:         "ConfigFileKeyRef select the configuration file.",
					MarkdownDescription: "ConfigFileKeyRef select the configuration file.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the configuration file.",
							MarkdownDescription: "Name is the name of the configuration file.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}
