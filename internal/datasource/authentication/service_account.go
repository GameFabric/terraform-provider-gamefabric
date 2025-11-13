package authentication

import (
	"context"
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
				Description:         "The service account username.",
				MarkdownDescription: "The service account username.",
				Optional:            true,
			},
			"email": schema.StringAttribute{
				Description:         "The service account email.",
				MarkdownDescription: "The service account email.",
				Optional:            true,
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

	var obj *authv1.ServiceAccount
	var err error

	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.AuthenticationV1Beta1().ServiceAccounts().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	case conv.IsKnown(config.Email):
		obj, err = r.findServiceAccountByEmail(ctx, config.Email.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Insufficient Information",
			"Either 'name' or 'email' must be provided to locate the service account.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Fetching Service Account",
			fmt.Sprintf("Failed to retrieve ServiceAccount: %v", err),
		)
		return
	}

	if obj == nil {
		resp.Diagnostics.AddError(
			"Service Account Not Found",
			"No service account was found matching the provided 'name' or 'email'.",
		)
		return
	}

	state := newServiceAccountModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccount) findServiceAccountByEmail(ctx context.Context, email string) (*authv1.ServiceAccount, error) {
	list, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, sa := range list.Items {
		if sa.Spec.Email == email {
			return &sa, nil
		}
	}

	return nil, fmt.Errorf("service account with email %q not found", email)
}
