package authentication

import (
	"context"
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &serviceAccountPassword{}
	_ ephemeral.EphemeralResourceWithConfigure = &serviceAccountPassword{}
)

// serviceAccountPassword implements the Terraform ephemeral resource for service account passwords.
type serviceAccountPassword struct {
	clientSet clientset.Interface
}

// NewServiceAccountPasswordEphemeralResource creates a new instance of the service account password ephemeral resource.
func NewServiceAccountPasswordEphemeralResource() ephemeral.EphemeralResource {
	return &serviceAccountPassword{}
}

func (r *serviceAccountPassword) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_password"
}

func (r *serviceAccountPassword) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Ephemeral resource for retrieving a service account password. The password is reset when the ephemeral resource is opened and is only available during the Terraform operation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"service_account": schema.StringAttribute{
				Required:    true,
				Description: "The name of the service account.",
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The password for the service account (reset on each use).",
			},
		},
	}
}

func (r *serviceAccountPassword) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *serviceAccountPassword) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config serviceAccountPasswordEphemeralModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify the service account exists.
	_, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().Get(ctx, config.ServiceAccount.ValueString(), metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Service Account",
			fmt.Sprintf("Could not read ServiceAccount: %s", err),
		)
		return
	}

	password, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().Reset(ctx, config.ServiceAccount.ValueString(), metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Error Opening Service Account Password", fmt.Sprintf("Could not reset ServiceAccount password: %s", err))
		return
	}

	resp.Diagnostics.AddWarning("Password reset dsiufghwofiwe", "The service account password has been reset.")

	config.ID = types.StringValue(config.ServiceAccount.ValueString())
	config.Password = types.StringValue(password)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}
