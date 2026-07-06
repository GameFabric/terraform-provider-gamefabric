package provisioning

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &allocator{}
	_ datasource.DataSourceWithConfigure = &allocator{}
)

type allocator struct {
	clientSet clientset.Interface
}

// NewAllocator creates a new allocator data source.
func NewAllocator() datasource.DataSource {
	return &allocator{}
}

// Metadata defines the data source type name.
func (r *allocator) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocator"
}

// Schema defines the schema for this data source.
func (r *allocator) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique name of the managed allocator.",
				MarkdownDescription: "The unique name of the managed allocator.",
				Required:            true,
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
				Description:         "The base URL of the registration service endpoint.",
				MarkdownDescription: "The base URL of the registration service endpoint.",
				Computed:            true,
			},
			"registry_active_token": schema.StringAttribute{
				Description:         "The active access token for the registration service endpoint.",
				MarkdownDescription: "The active access token for the registration service endpoint.",
				Computed:            true,
				Sensitive:           true,
			},
			"registry_tokens": schema.ListAttribute{
				Description:         "The ordered list of access tokens for the registration service endpoint. Deprecated tokens are phased out. The last token in the list is the most recent and exposed as active token.",
				MarkdownDescription: "The ordered list of access tokens for the registration service endpoint. Deprecated tokens are phased out. The last token in the list is the most recent and exposed as active token.",
				Computed:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *allocator) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *allocator) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config allocatorModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.ProvisioningV1Beta1().Allocators().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Allocator Not Found",
				fmt.Sprintf("Allocator %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting Allocator",
			fmt.Sprintf("Could not get Allocator %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newAllocatorModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
