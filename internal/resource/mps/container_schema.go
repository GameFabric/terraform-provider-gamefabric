package mps

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ContainersAttributes returns the schema attributes for containers.
func ContainersAttributes() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "Name is the name of the container. The primary gameserver container should be named `default`",
				MarkdownDescription: "Name is the name of the container. The primary gameserver container should be named `default`",
				Required:            true,
			},
			"image_ref": schema.SingleNestedAttribute{
				Description: "ImageRef is the GameFabric container image reference to run. You should use the `image_ref` attribute of the `gamefabric_image` datasource to configure this attribute.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description:         "Name is the name of the GameFabric image.",
						MarkdownDescription: "Name is the name of the GameFabric image.",
						Required:            true,
						Validators: []validator.String{
							validators.NameValidator{},
						},
					},
					"branch": schema.StringAttribute{
						Description:         "Branch of the GameFabric image.",
						MarkdownDescription: "Branch of the GameFabric image.",
						Required:            true,
						Validators: []validator.String{
							validators.NameValidator{},
						},
					},
				},
			},
			"command": schema.ListAttribute{
				Description:         "Command is the entrypoint array. This is not executed within a shell.",
				MarkdownDescription: "Command is the entrypoint array. This is not executed within a shell.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"args": schema.ListAttribute{
				Description:         "Args are arguments to the entrypoint.",
				MarkdownDescription: "Args are arguments to the entrypoint.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"resources": schema.SingleNestedAttribute{
				Description:         "Resources describes the compute resource requirements.",
				MarkdownDescription: "Resources describes the compute resource requirements. See the <a href=\"https://docs.gamefabric.com/multiplayer-servers/multiplayer-services/resource-management\">GameFabric documentation</a> for more details on how to configure resource requests and limits.",
				Optional:            true,
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
								},
							},
							"memory": schema.StringAttribute{
								Description:         "Memory limit.",
								MarkdownDescription: "Memory limit.",
								Optional:            true,
								Validators: []validator.String{
									validators.QuantityValidator{},
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
								},
							},
							"memory": schema.StringAttribute{
								Description:         "Memory request.",
								MarkdownDescription: "Memory request.",
								Optional:            true,
								Validators: []validator.String{
									validators.QuantityValidator{},
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
				NestedObject: schema.NestedAttributeObject{
					Attributes: core.EnvVarAttributes(),
				},
			},
			"ports": schema.ListNestedAttribute{
				Description:         "Ports are the ports to expose from the container.",
				MarkdownDescription: "Ports are the ports to expose from the container.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the port. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
							MarkdownDescription: "Name is the name of the port. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"policy": schema.StringAttribute{
							Description:         "Policy defines how the host port is populated. Dynamic (default) allocates a free host port and maps it to the container_port (required). The gameserver must report the external port (obtained via Agones SDK) to backends for client connections. Passthrough dynamically allocates a host port and sets container_port to match it. The gameserver must discover this port via Agones SDK and listen on it.",
							MarkdownDescription: "Policy defines how the host port is populated. `Dynamic` (default) allocates a free host port and maps it to the `container_port` (required). The gameserver must report the external port (obtained via Agones SDK) to backends for client connections. `Passthrough` dynamically allocates a host port and sets `container_port` to match it. The gameserver must discover this port via Agones SDK and listen on it.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("Dynamic", "Passthrough"),
							},
						},
						"container_port": schema.Int32Attribute{
							Description:         "ContainerPort is the port that is being opened on the specified container&#39;s process.",
							MarkdownDescription: "ContainerPort is the port that is being opened on the specified container&#39;s process.",
							Optional:            true,
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
						"protocol": schema.StringAttribute{
							Description:         "Protocol is the network protocol being used. Defaults to UDP. TCP is the other option.",
							MarkdownDescription: "Protocol is the network protocol being used. Defaults to UDP. TCP is the other option.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("UDP", "TCP"), // GameFabric does not support TCPUDP.
							},
						},
						"protection_protocol": schema.StringAttribute{
							Description:         "ProtectionProtocol is the optional name of the protection protocol being used.",
							MarkdownDescription: "ProtectionProtocol is the optional name of the protection protocol being used.",
							Optional:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
					},
				},
			},
			"volume_mounts": schema.ListNestedAttribute{
				Description:         "VolumeMounts are the volumes to mount into the container&#39;s filesystem.",
				MarkdownDescription: "VolumeMounts are the volumes to mount into the container&#39;s filesystem.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the volume.",
							MarkdownDescription: "Name is the name of the volume.",
							Optional:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"mount_path": schema.StringAttribute{
							Description:         "Path within the container at which the volume should be mounted.",
							MarkdownDescription: "Path within the container at which the volume should be mounted.",
							Optional:            true,
						},
						"sub_path": schema.StringAttribute{
							Description:         "Path within the volume from which the container's volume should be mounted. Defaults to empty string (volume's root).",
							MarkdownDescription: "Path within the volume from which the container's volume should be mounted.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sub_path_expr")),
							},
						},
						"sub_path_expr": schema.StringAttribute{
							Description:         "Expanded path within the volume from which the container's volume should be mounted.",
							MarkdownDescription: "Expanded path within the volume from which the container's volume should be mounted.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sub_path")),
							},
						},
					},
				},
			},
			"config_files": schema.ListNestedAttribute{
				Description:         "ConfigFiles is a list of configuration files to mount into the containers filesystem.",
				MarkdownDescription: "ConfigFiles is a list of configuration files to mount into the containers filesystem.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the configuration file.",
							MarkdownDescription: "Name is the name of the configuration file.",
							Required:            true,
						},
						"mount_path": schema.StringAttribute{
							Description:         "MountPath is the path to mount the configuration file on.",
							MarkdownDescription: "MountPath is the path to mount the configuration file on.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}
