package authentication

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &serviceAccountPassword{}
	_ resource.ResourceWithConfigure = &serviceAccountPassword{}
)

// serviceAccountPassword implements the Terraform resource for service account passwords.
type serviceAccountPassword struct {
	clientSet clientset.Interface
}

// NewServiceAccountPasswordResource creates a new instance of the service account password resource.
func NewServiceAccountPasswordResource() resource.Resource {
	return &serviceAccountPassword{}
}

func (r *serviceAccountPassword) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_password"
}

func (r *serviceAccountPassword) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_account": schema.StringAttribute{
				Required:    true,
				Description: "The name of the service account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The password for the service account (read-only, reset on creation or update).",
			},
		},
	}
}

func (r *serviceAccountPassword) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceAccountPassword) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountPasswordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().Reset(ctx, plan.ServiceAccount.ValueString(), metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Service Account Password", fmt.Sprintf("Could not reset ServiceAccount password: %s", err))
		return
	}

	plan.ID = types.StringValue(plan.ServiceAccount.ValueString())
	plan.Password = types.StringValue(password)

	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceAccountPassword) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountPasswordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().Get(ctx, state.ServiceAccount.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Service Account Password",
				fmt.Sprintf("Could not read ServiceAccount: %s", err),
			)
		}
		return
	}

	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccountPassword) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountPasswordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Deleting a password resource just removes it from state; no actual deletion happens on the server.
	// The password itself remains valid; we're just stopping tracking this resource.
}

func (r *serviceAccountPassword) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This method is required by the resource.Resource interface but should never be called
	// because the only mutable field (service_account) has RequiresReplace(), which forces
	// a destroy/recreate instead of an update.
	//
	// If this is somehow called, we just read the current state back.
	var state serviceAccountPasswordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
