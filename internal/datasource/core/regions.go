package core

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
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
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter regions.",
				MarkdownDescription: "A map of keys and values that is used to filter regions.",
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
						"display_name": schema.StringAttribute{
							Description:         "The display name of the region.",
							MarkdownDescription: "The display name of the region.",
							Computed:            true,
						},
						"environment": schema.StringAttribute{
							Description:         "The name of the environment the object belongs to.",
							MarkdownDescription: "The name of the environment the object belongs to.",
							Computed:            true,
						},
						"types": schema.MapNestedAttribute{
							Description:         "Types defines the types on infrastructure available in the region.",
							MarkdownDescription: "Types defines the types on infrastructure available in the region.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
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

	labelSelector := conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() })
	list, err := r.clientSet.CoreV1().Regions(config.Environment.ValueString()).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Regions",
			fmt.Sprintf("Could not get Regions: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b corev1.Region) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newRegionsModel(list.Items)
	state.Environment = config.Environment
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
