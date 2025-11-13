package protection

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
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
)

var (
	_ datasource.DataSource              = &protocol{}
	_ datasource.DataSourceWithConfigure = &protocol{}
)

type protocol struct {
	clientSet clientset.Interface
}

// NewProtocol creates a new protocol data source.
func NewProtocol() datasource.DataSource {
	return &protocol{}
}

// Metadata defines the data source type name.
func (r *protocol) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_protocol"
}

// Schema defines the schema for this data source.
func (r *protocol) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique protection protocol object name.",
				MarkdownDescription: "The unique protection protocol object name.",
				Optional:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("display_name")),
				},
			},
			"display_name": schema.StringAttribute{
				Description:         "The user-friendly name of the protection protocol.",
				MarkdownDescription: "The user-friendly name of the protection protocol.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
				},
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
	}
}

// Configure prepares the struct.
func (r *protocol) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *protocol) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config protocolModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		obj *protectionv1.Protocol
		err error
	)
	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.ProtectionV1().Protocols().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Protocol Not Found",
					fmt.Sprintf("Protocol %q was not found.", config.Name.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Getting Protocol",
				fmt.Sprintf("Could not get Protocol %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	case conv.IsKnown(config.DisplayName):
		list, err := r.clientSet.ProtectionV1().Protocols().List(ctx, metav1.ListOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Protocol",
				fmt.Sprintf("Could not get Protocol with display name %q: %v", config.DisplayName.ValueString(), err),
			)
			return
		}
		for _, item := range list.Items {
			if item.Spec.DisplayName != config.DisplayName.ValueString() {
				continue
			}
			if obj != nil {
				resp.Diagnostics.AddError(
					"Multiple Protocols Found",
					fmt.Sprintf("Multiple Protocols found with display name %q", config.DisplayName.ValueString()),
				)
				return
			}
			obj = &item
		}
		if obj == nil {
			resp.Diagnostics.AddError(
				"Protocol Not Found",
				fmt.Sprintf("Protocol with display name %q was not found.", config.DisplayName.ValueString()),
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

	state := newProtocolModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
