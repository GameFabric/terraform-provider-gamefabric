package mps

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ContainersAttributes returns the schema attributes for containers.
//
//nolint:maintidx
func ContainersAttributes(val validators.GameFabricValidator, pathPrefix string) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "Name is the name of the container. The primary game server container should be named `default`",
				MarkdownDescription: "Name is the name of the container. The primary game server container should be named `default`",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					validators.GFFieldString(val, pathPrefix+".name"),
				},
			},
			"image_ref": schema.SingleNestedAttribute{
				Description: "The reference to the image to run your container with. You should use the `image_ref` attribute of the `gamefabric_image` datasource to configure this attribute.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description:         "Name is the name of the GameFabric image.",
						MarkdownDescription: "Name is the name of the GameFabric image.",
						Required:            true,
						Validators: []validator.String{
							validators.NameValidator{},
							validators.GFFieldString(val, pathPrefix+".image"),
						},
					},
					"branch": schema.StringAttribute{
						Description:         "Branch of the GameFabric image.",
						MarkdownDescription: "Branch of the GameFabric image.",
						Required:            true,
						Validators: []validator.String{
							validators.NameValidator{},
							validators.GFFieldString(val, pathPrefix+".branch"),
						},
					},
				},
			},
			"command": schema.ListAttribute{
				Description:         "Command is the entrypoint array. This is not executed within a shell.",
				MarkdownDescription: "Command is the entrypoint array. This is not executed within a shell.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(validators.GFFieldString(val, pathPrefix+".command[?]")),
				},
			},
			"args": schema.ListAttribute{
				Description:         "Args are arguments to the entrypoint.",
				MarkdownDescription: "Args are arguments to the entrypoint.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(validators.GFFieldString(val, pathPrefix+".args[?]")),
				},
			},
			"resources": schema.SingleNestedAttribute{
				Description:         "Resources describes the compute resource requirements.",
				MarkdownDescription: "Resources describes the compute resource requirements. See the <a href=\"https://docs.gamefabric.com/multiplayer-servers/multiplayer-services/resource-management\">GameFabric documentation</a> for more details on how to configure resource requests and limits.",
				Optional:            true,
				Validators: []validator.Object{
					validators.GFFieldObject(val, pathPrefix+".resources"),
				},
				Attributes: map[string]schema.Attribute{
					"limits": schema.SingleNestedAttribute{
						Description:         "Limits describes the maximum amount of compute resources allowed.",
						MarkdownDescription: "Limits describes the maximum amount of compute resources allowed.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"cpu": schema.StringAttribute{
								Description:         "CPU limit.",
								MarkdownDescription: "CPU limit.",
								Optional:            true,
								Validators: []validator.String{
									validators.QuantityValidator{},
									validators.GFFieldString(val, pathPrefix+".resources.limits.cpu"),
								},
							},
							"memory": schema.StringAttribute{
								Description:         "Memory limit.",
								MarkdownDescription: "Memory limit.",
								Optional:            true,
								Validators: []validator.String{
									validators.QuantityValidator{},
									validators.GFFieldString(val, pathPrefix+".resources.limits.memory"),
								},
							},
						},
					},
					"requests": schema.SingleNestedAttribute{
						Description:         "Requests describes the minimum amount of compute resources required.",
						MarkdownDescription: "Requests describes the minimum amount of compute resources required.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"cpu": schema.StringAttribute{
								Description:         "CPU request.",
								MarkdownDescription: "CPU request.",
								Optional:            true,
								Validators: []validator.String{
									validators.QuantityValidator{},
									validators.GFFieldString(val, pathPrefix+".resources.requests.cpu"),
								},
							},
							"memory": schema.StringAttribute{
								Description:         "Memory request.",
								MarkdownDescription: "Memory request.",
								Optional:            true,
								Validators: []validator.String{
									validators.QuantityValidator{},
									validators.GFFieldString(val, pathPrefix+".resources.requests.memory"),
								},
							},
						},
					},
				},
			},
			"envs": schema.ListNestedAttribute{
				Description:         "Envs is a list of environment variables to set on all containers in this Armada.",
				MarkdownDescription: "Envs is a list of environment variables to set on all containers in this Armada.",
				Optional:            true,
				Validators: []validator.List{
					validators.GFFieldList(val, pathPrefix+".env"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: core.EnvVarAttributes(val, pathPrefix+".env[?]"),
				},
			},
			"ports": schema.ListNestedAttribute{
				Description:         "Ports are the ports to expose from the container.",
				MarkdownDescription: "Ports are the ports to expose from the container.",
				Optional:            true,
				Validators: []validator.List{
					validators.GFFieldList(val, pathPrefix+".ports"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the port. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
							MarkdownDescription: "Name is the name of the port. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
								validators.GFFieldString(val, pathPrefix+".ports[?].name"),
							},
						},
						"policy": schema.StringAttribute{
							Description:         "Policy defines how the host port is populated. Dynamic (default) allocates a free host port and maps it to the container_port (required). Passthrough dynamically allocates a host port and sets container_port to match it. Metric enables Prometheus scraping: it injects the g8c.io/gameserver-scrape label and g8c.io/metrics-endpoints annotation on the GameServer template, and is stored as policy None in the GameFabric API. Metric ports require protocol=TCP and a container_port.",
							MarkdownDescription: "Policy defines how the host port is populated. `Dynamic` (default) allocates a free host port and maps it to the `container_port` (required). `Passthrough` dynamically allocates a host port and sets `container_port` to match it. `Metric` enables Prometheus scraping: it injects the `g8c.io/gameserver-scrape` label and `g8c.io/metrics-endpoints` annotation on the GameServer template, and is stored as policy `None` in the GameFabric API. `Metric` ports require `protocol = \"TCP\"` and a `container_port`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("Dynamic", "Passthrough", "Metric"),
								// GFFieldString is intentionally omitted for policy: "Metric" is a
								// TFP-only value that gets translated to "None" before hitting the API.
								metricPolicyValidator{},
							},
						},
						"container_port": schema.Int32Attribute{
							Description:         "ContainerPort is the port that is being opened on the specified container&#39;s process.",
							MarkdownDescription: "ContainerPort is the port that is being opened on the specified container&#39;s process.",
							Optional:            true,
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
								validators.GFFieldInt32(val, pathPrefix+".ports[?].containerPort"),
							},
						},
						"protocol": schema.StringAttribute{
							Description:         "Protocol is the network protocol being used. Defaults to UDP. TCP is the other option.",
							MarkdownDescription: "Protocol is the network protocol being used. Defaults to UDP. TCP is the other option.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("UDP", "TCP"), // GameFabric does not support TCPUDP.
								validators.GFFieldString(val, pathPrefix+".ports[?].protocol"),
							},
						},
						"protection_protocol": schema.StringAttribute{
							Description:         "ProtectionProtocol is the optional name of the protection protocol being used.",
							MarkdownDescription: "ProtectionProtocol is the optional name of the protection protocol being used.",
							Optional:            true,
							Validators: []validator.String{
								validators.NameValidator{},
								validators.GFFieldString(val, pathPrefix+".ports[?].protectionProtocol"),
							},
						},
						"path": schema.StringAttribute{
							Description:         "Path is the HTTP path for Prometheus metrics scraping. Only valid when policy is Metric. Defaults to /metrics when policy is Metric and path is omitted.",
							MarkdownDescription: "Path is the HTTP path for Prometheus metrics scraping. Only valid when `policy` is `Metric`. Defaults to `/metrics` when `policy` is `Metric` and `path` is omitted.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								pathOnlyForMetricValidator{},
								noCommaValidator{},
							},
						},
					},
				},
			},
			"volume_mounts": schema.ListNestedAttribute{
				Description:         "VolumeMounts are the volumes to mount into the container&#39;s filesystem.",
				MarkdownDescription: "VolumeMounts are the volumes to mount into the container&#39;s filesystem.",
				Optional:            true,
				Validators: []validator.List{
					validators.GFFieldList(val, pathPrefix+".volumeMounts"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the volume.",
							MarkdownDescription: "Name is the name of the volume.",
							Optional:            true,
							Validators: []validator.String{
								validators.NameValidator{},
								validators.GFFieldString(val, pathPrefix+".volumeMounts[?].name"),
							},
						},
						"mount_path": schema.StringAttribute{
							Description:         "Path within the container at which the volume should be mounted.",
							MarkdownDescription: "Path within the container at which the volume should be mounted.",
							Optional:            true,
							Validators: []validator.String{
								validators.GFFieldString(val, pathPrefix+".volumeMounts[?].mountPath"),
							},
						},
						"sub_path": schema.StringAttribute{
							Description:         "Path within the volume from which the container's volume should be mounted. Defaults to empty string (volume's root).",
							MarkdownDescription: "Path within the volume from which the container's volume should be mounted.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sub_path_expr")),
								validators.GFFieldString(val, pathPrefix+".volumeMounts[?].subPath"),
							},
						},
						"sub_path_expr": schema.StringAttribute{
							Description:         "Expanded path within the volume from which the container's volume should be mounted.",
							MarkdownDescription: "Expanded path within the volume from which the container's volume should be mounted.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sub_path")),
								validators.GFFieldString(val, pathPrefix+".volumeMounts[?].subPathExpr"),
							},
						},
					},
				},
			},
			"config_files": schema.ListNestedAttribute{
				Description:         "ConfigFiles is a list of configuration files to mount into the containers filesystem.",
				MarkdownDescription: "ConfigFiles is a list of configuration files to mount into the containers filesystem.",
				Optional:            true,
				Validators: []validator.List{
					validators.GFFieldList(val, pathPrefix+".configFiles"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the configuration file.",
							MarkdownDescription: "Name is the name of the configuration file.",
							Required:            true,
							Validators: []validator.String{
								validators.GFFieldString(val, pathPrefix+".configFiles[?].name"),
							},
						},
						"mount_path": schema.StringAttribute{
							Description:         "MountPath is the path to mount the configuration file on.",
							MarkdownDescription: "MountPath is the path to mount the configuration file on.",
							Required:            true,
							Validators: []validator.String{
								validators.GFFieldString(val, pathPrefix+".configFiles[?].mountPath"),
							},
						},
					},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Description:         "Secrets is a list of secrets to mount into the containers filesystem.",
				MarkdownDescription: "Secrets is a list of secrets to mount into the containers filesystem.",
				Optional:            true,
				Validators: []validator.List{
					validators.GFFieldList(val, pathPrefix+".secrets"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the secret.",
							MarkdownDescription: "Name is the name of the secret.",
							Required:            true,
							Validators: []validator.String{
								validators.GFFieldString(val, pathPrefix+".secrets[?].name"),
							},
						},
						"mount_path": schema.StringAttribute{
							Description:         "MountPath is the path to mount the secret on.",
							MarkdownDescription: "MountPath is the path to mount the secret on.",
							Required:            true,
							Validators: []validator.String{
								validators.GFFieldString(val, pathPrefix+".secrets[?].mountPath"),
							},
						},
					},
				},
			},
		},
	}
}
