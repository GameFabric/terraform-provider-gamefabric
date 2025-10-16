package protection

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &protocols{}
	_ datasource.DataSourceWithConfigure = &protocols{}
)

type protocols struct {
	clientSet clientset.Interface
}

// NewProtocols returns a new instance of the protocols data source.
func NewProtocols() datasource.DataSource {
	return &protocols{}
}

func (r *protocols) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_protocols"
}

// Schema defines the schema for this data source.
func (r *protocols) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protocols": schema.ListNestedAttribute{
				Description:         "A list of protocols.",
				MarkdownDescription: "A list of protocols.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name within its scope.",
							MarkdownDescription: "The unique object name within its scope.",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							Description:         "DisplayName is the user-friendly name of a region.",
							MarkdownDescription: "DisplayName is the user-friendly name of a region.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "The description of the protocol.",
							MarkdownDescription: "The description of the protocol.",
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							Description:         "The network protocol used by this protocol.",
							MarkdownDescription: "The network protocol used by this protocol.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *protocols) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *protocols) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config protocolsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.ProtectionV1().Protocols().List(ctx, metav1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Protocols",
			fmt.Sprintf("Could not get Protocols: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b protectionv1.Protocol) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newProtocolsModel(list.Items)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
