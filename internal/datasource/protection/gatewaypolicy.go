package protection

import (
	"context"
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &gatewayPolicy{}
	_ datasource.DataSourceWithConfigure = &gatewayPolicy{}
)

type gatewayPolicy struct {
	clientSet clientset.Interface
}

// NewGatewayPolicy creates a new gateway policy data source.
func NewGatewayPolicy() datasource.DataSource {
	return &gatewayPolicy{}
}

// Metadata defines the data source type name.
func (r *gatewayPolicy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_gatewaypolicy"
}

// Schema defines the schema for this data source.
func (r *gatewayPolicy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope.",
				MarkdownDescription: "The unique object name within its scope.",
				Optional:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("display_name")),
				},
			},
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is friendly name of the gateway policy.",
				MarkdownDescription: "DisplayName is friendly name of the gateway policy.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"labels": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"description": schema.StringAttribute{
				Description:         "Description is the description of the gateway policy.",
				MarkdownDescription: "Description is the description of the gateway policy.",
				Computed:            true,
			},
			"destination_cidrs": schema.ListAttribute{
				Description:         "The CIDRs that should use the gateway for outbound traffic, rather than the game server node.",
				MarkdownDescription: "The CIDRs that should use the gateway for outbound traffic, rather than the game server node.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *gatewayPolicy) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *gatewayPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config gatewayPolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		obj *protectionv1.GatewayPolicy
		err error
	)
	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.ProtectionV1().GatewayPolicies().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Gateway Policy",
				fmt.Sprintf("Could not get Gateway Policy %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	case conv.IsKnown(config.DisplayName):
		list, err := r.clientSet.ProtectionV1().GatewayPolicies().List(ctx, metav1.ListOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Gateway Policy",
				fmt.Sprintf("Could not get Gateway Policy with display name %q: %v", config.DisplayName.ValueString(), err),
			)
			return
		}
		for _, item := range list.Items {
			if item.Spec.DisplayName != config.DisplayName.ValueString() {
				continue
			}
			if obj != nil {
				resp.Diagnostics.AddError(
					"Multiple Gateway Policies Found",
					fmt.Sprintf("Multiple Gateway Policies found with display name %q", config.DisplayName.ValueString()),
				)
				return
			}
			obj = &item
		}
		if obj == nil {
			resp.Diagnostics.AddError(
				"Gateway Policy Not Found",
				fmt.Sprintf("Gateway Policy with display name %q was not found.", config.DisplayName.ValueString()),
			)
			return
		}
	default:
		resp.Diagnostics.AddError(
			"Insufficient Information",
			"Either name or display_name must be specified.",
		)
		return
	}

	state := newGatewayPolicyModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
