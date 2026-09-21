package provisioning

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &allocators{}
	_ datasource.DataSourceWithConfigure = &allocators{}
)

type allocators struct {
	clientSet clientset.Interface
}

// NewAllocators returns a new instance of the allocators data source.
func NewAllocators() datasource.DataSource {
	return &allocators{}
}

// Metadata defines the data source type name.
func (r *allocators) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocators"
}

// Schema defines the schema for this data source.
func (r *allocators) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source for a list of Allocators.",
		MarkdownDescription: "Data source for a list of Allocators.",
		Attributes: map[string]schema.Attribute{
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter allocators. Only items with all specified labels (exact matches) will be returned.",
				MarkdownDescription: "A map of keys and values that is used to filter allocators. Only items with all specified labels (exact matches) will be returned.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"allocators": schema.ListNestedAttribute{
				Description:         "The allocators that match the label filter.",
				MarkdownDescription: "The allocators that match the label filter.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique name of the managed allocator.",
							MarkdownDescription: "The unique name of the managed allocator.",
							Computed:            true,
						},
						"region": schema.StringAttribute{
							Description:         "The deployment region of the allocator. This acts as a geographical indicator but might differ from custom GameFabric regions.",
							MarkdownDescription: "The deployment region of the allocator. This acts as a geographical indicator but might differ from custom GameFabric regions.",
							Computed:            true,
						},
						"rate_limit_qps": schema.Int64Attribute{
							Description:         "The maximum number of queries per second.",
							MarkdownDescription: "The maximum number of queries per second.",
							Computed:            true,
						},
						"rate_limit_burst": schema.Int64Attribute{
							Description:         "The maximum number of queries that can be served at once before the rate limit kicks in. Excess capacity builds up during idle periods, up to this limit.",
							MarkdownDescription: "The maximum number of queries that can be served at once before the rate limit kicks in. Excess capacity builds up during idle periods, up to this limit.",
							Computed:            true,
						},
						"allocation_url": schema.StringAttribute{
							Description:         "The base URL of the allocation service endpoint. The endpoint returns a game server that matches the requested attributes.",
							MarkdownDescription: "The base URL of the allocation service endpoint. The endpoint returns a game server that matches the requested attributes.",
							Computed:            true,
						},
						"allocation_active_token": schema.StringAttribute{
							Description:         "The active access token for the allocation service endpoint.",
							MarkdownDescription: "The active access token for the allocation service endpoint.",
							Computed:            true,
							Sensitive:           true,
						},
						"allocation_tokens": schema.ListAttribute{
							Description:         "The ordered list of access tokens for the allocation service endpoint. Deprecated tokens are phased out. The last token in the list is the most recent and exposed as active token.",
							MarkdownDescription: "The ordered list of access tokens for the allocation service endpoint. Deprecated tokens are phased out. The last token in the list is the most recent and exposed as active token.",
							Computed:            true,
							Sensitive:           true,
							ElementType:         types.StringType,
						},
						"registry_url": schema.StringAttribute{
							Description:         "The base URL of the registry service endpoint.",
							MarkdownDescription: "The base URL of the registry service endpoint.",
							Computed:            true,
						},
						"registry_active_token": schema.StringAttribute{
							Description:         "The active access token for the registry service endpoint.",
							MarkdownDescription: "The active access token for the registry service endpoint.",
							Computed:            true,
							Sensitive:           true,
						},
						"registry_tokens": schema.ListAttribute{
							Description:         "The ordered list of access tokens for the registry service endpoint. Deprecated tokens are phased out. The last token in the list is the most recent and exposed as active token.",
							MarkdownDescription: "The ordered list of access tokens for the registry service endpoint. Deprecated tokens are phased out. The last token in the list is the most recent and exposed as active token.",
							Computed:            true,
							Sensitive:           true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *allocators) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *allocators) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config allocatorsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.ProvisioningV1Beta1().Allocators().List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Allocators",
			fmt.Sprintf("Could not get Allocators: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b provisioningv1beta1.Allocator) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newAllocatorsModel(list.Items)
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
