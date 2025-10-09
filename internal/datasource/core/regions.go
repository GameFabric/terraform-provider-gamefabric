package core

import (
	"context"
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apicore/labels"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &regions{}
	_ datasource.DataSourceWithConfigure = &regions{}
)

type regions struct {
	clientSet clientset.Interface
}

// NewRegions creates a new regions data source.
func NewRegions() datasource.DataSource {
	return &regions{}
}

// Metadata defines the data source type name.
func (r *regions) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

// Schema defines the schema for this data source.
func (r *regions) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the object belongs to.",
				MarkdownDescription: "The name of the environment the object belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
			},
			"labels_filter": schema.MapAttribute{
				Description:         "Locations defines the locations for a type.",
				MarkdownDescription: "Locations defines the locations for a type.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"regions": schema.ListNestedAttribute{
				Description:         "Regions is a list of regions in the environment that match the labels filter.",
				MarkdownDescription: "Regions is a list of regions in the environment that match the labels filter.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name within its scope.",
							MarkdownDescription: "The unique object name within its scope.",
							Computed:            true,
						},
						"environment": schema.StringAttribute{
							Description:         "The name of the environment the object belongs to.",
							MarkdownDescription: "The name of the environment the object belongs to.",
							Computed:            true,
						},
						"labels": schema.MapAttribute{
							Description:         "A map of keys and values that can be used to organize and categorize objects.",
							MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"display_name": schema.StringAttribute{
							Description:         "DisplayName is the user-friendly name of a region.",
							MarkdownDescription: "DisplayName is the user-friendly name of a region.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description is the optional description of the region.",
							MarkdownDescription: "Description is the optional description of the region.",
							Computed:            true,
						},
						"types": schema.MapNestedAttribute{
							Description:         "Types defines the types on infrastructure available in the region.",
							MarkdownDescription: "Types defines the types on infrastructure available in the region.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"locations": schema.ListAttribute{
										Description:         "Locations defines the locations for a type.",
										MarkdownDescription: "Locations defines the locations for a type.",
										Computed:            true,
										ElementType:         types.StringType,
									},
									"env": schema.ListNestedAttribute{
										Description:         "Env is a list of environment variables to set on all containers in this region.",
										MarkdownDescription: "Env is a list of environment variables to set on all containers in this region.",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
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
											},
										},
									},
									"scheduling": schema.StringAttribute{
										Description:         "Scheduling strategy. Defaults to &#34;Packed&#34;",
										MarkdownDescription: "Scheduling strategy. Defaults to &#34;Packed&#34;",
										Computed:            true,
									},
									"cpu": schema.StringAttribute{
										Description:         "CPU is the CPU limit for the region type.",
										MarkdownDescription: "CPU is the CPU limit for the region type.",
										Computed:            true,
									},
									"memory": schema.StringAttribute{
										Description:         "Memory is the memory limit for the region type.",
										MarkdownDescription: "Memory is the memory limit for the region type.",
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *regions) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	procCtx, ok := req.ProviderData.(*provcontext.Context)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *provider.Context, got %T", req.ProviderData),
		)
		return
	}

	r.clientSet = procCtx.ClientSet
}

func (r *regions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config regionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labelSelector := labels.SelectorFromSet(conv.ForEachMapItem(config.LabelsFilter, func(item types.String) string { return item.ValueString() }))
	list, err := r.clientSet.CoreV1().Regions(config.Environment.ValueString()).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Region",
			fmt.Sprintf("Could not list Regions: %v", err),
		)
		return
	}

	for _, item := range list.Items {
		config.Regions = append(config.Regions, newRegionModel(&item))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type regionsModel struct {
	Environment  types.String            `tfsdk:"environment"`
	LabelsFilter map[string]types.String `tfsdk:"labels_filter"`
	Regions      []regionModel           `tfsdk:"regions"`
}
