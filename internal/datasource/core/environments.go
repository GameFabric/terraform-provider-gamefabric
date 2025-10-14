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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &environments{}
	_ datasource.DataSourceWithConfigure = &environments{}
)

type environments struct {
	clientSet clientset.Interface
}

// NewEnvironments returns a new instance of the environments data source.
func NewEnvironments() datasource.DataSource {
	return &environments{}
}

func (r *environments) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

// Schema defines the schema for this data source.
func (r *environments) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter environments.",
				MarkdownDescription: "A map of keys and values that is used to filter environments.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"environments": schema.ListNestedAttribute{
				Description:         "A list of environments matching the labels filter.",
				MarkdownDescription: "A list of environments matching the labels filter.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name within its scope.",
							MarkdownDescription: "The unique object name within its scope.",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							Description:         "DisplayName is friendly name of the environment.",
							MarkdownDescription: "DisplayName is friendly name of the environment.",
							Computed:            true,
						},
						"labels": schema.MapAttribute{
							Description:         "A map of keys and values that can be used to organize and categorize objects.",
							MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"description": schema.StringAttribute{
							Description:         "Description is the description of the environment.",
							MarkdownDescription: "Description is the description of the environment.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *environments) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *environments) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config environmentsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.CoreV1().Environments().List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Environments",
			fmt.Sprintf("Could not get Environments: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b corev1.Environment) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newEnvironmentsModel(list.Items)
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
