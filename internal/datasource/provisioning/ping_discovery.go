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
	_ datasource.DataSource              = &pingDiscovery{}
	_ datasource.DataSourceWithConfigure = &pingDiscovery{}
)

type pingDiscovery struct {
	clientSet clientset.Interface
}

// NewPingDiscovery creates a new ping discovery data source.
func NewPingDiscovery() datasource.DataSource {
	return &pingDiscovery{}
}

// Metadata defines the data source type name.
func (r *pingDiscovery) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ping_discovery"
}

// Schema defines the schema for this data source.
func (r *pingDiscovery) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique ping discovery name.",
				MarkdownDescription: "The unique ping discovery name.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				Description:         "The base URL of the ping discovery endpoint.",
				MarkdownDescription: "The base URL of the ping discovery endpoint.",
				Computed:            true,
			},
			"active_token": schema.StringAttribute{
				Description:         "The active access token for the ping discovery endpoint.",
				MarkdownDescription: "The active access token for the ping discovery endpoint.",
				Computed:            true,
				Sensitive:           true,
			},
			"tokens": schema.ListAttribute{
				Description:         "The ordered list of access tokens for the ping discovery endpoint.",
				MarkdownDescription: "The ordered list of access tokens for the ping discovery endpoint.",
				Computed:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *pingDiscovery) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *pingDiscovery) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pingDiscoveryModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.ProvisioningV1Beta1().PingDiscoveries().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Ping Discovery Not Found",
				fmt.Sprintf("Ping Discovery %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting Ping Discovery",
			fmt.Sprintf("Could not get Ping Discovery %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newPingDiscoveryModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
