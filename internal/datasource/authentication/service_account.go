package authentication

import (
	"context"
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &serviceAccount{}
	_ datasource.DataSourceWithConfigure = &serviceAccount{}
)

type serviceAccount struct {
	clientSet clientset.Interface
}

// NewServiceAccount creates a new service account data source.
func NewServiceAccount() datasource.DataSource { return &serviceAccount{} }

// Metadata defines the data source type name.
func (r *serviceAccount) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

// Schema defines the schema for this data source.
func (r *serviceAccount) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name.",
				MarkdownDescription: "The unique object name.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"username": schema.StringAttribute{
				Description:         "The service account username.",
				MarkdownDescription: "The service account username.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				Description:         "The service account email.",
				MarkdownDescription: "The service account email.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				Description:         "The service account state.",
				MarkdownDescription: "The service account state.",
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				Description:         "Key-value labels for the service account.",
				MarkdownDescription: "Key-value labels for the service account.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *serviceAccount) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves the ServiceAccount and sets the state.
func (r *serviceAccount) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceAccountModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Service Account",
			fmt.Sprintf("Could not get ServiceAccount %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newServiceAccountModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
