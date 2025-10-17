package protection

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &gatewayPolicies{}
	_ datasource.DataSourceWithConfigure = &gatewayPolicies{}
)

type gatewayPolicies struct {
	clientSet clientset.Interface
}

// NewGatewayPolicies returns a new instance of the gateway policies data source.
func NewGatewayPolicies() datasource.DataSource {
	return &gatewayPolicies{}
}

func (r *gatewayPolicies) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_gatewaypolicies"
}

// Schema defines the schema for this data source.
func (r *gatewayPolicies) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter gateway policies.",
				MarkdownDescription: "A map of keys and values that is used to filter gateway policies.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"names": schema.ListAttribute{
				Description:         "The gateway policies matching the filters.",
				MarkdownDescription: "The gateway policies matching the filters.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *gatewayPolicies) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *gatewayPolicies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config gatewayPoliciesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.ProtectionV1().GatewayPolicies().List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Gateway Policies",
			fmt.Sprintf("Could not get Gateway Policies: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b protectionv1.GatewayPolicy) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newGatewayPoliciesModel(list.Items)
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
