package rbac

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/rbac/rolebinding"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &roleBinding{}
	_ resource.ResourceWithConfigure   = &roleBinding{}
	_ resource.ResourceWithImportState = &roleBinding{}
)

var roleBindingValidator = validators.NewGameFabricValidator[*rbacv1.RoleBinding, roleBindingModel](func() validators.StoreValidator {
	storage, _ := rolebinding.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

type roleBinding struct {
	clientSet clientset.Interface
}

// NewRoleBinding returns a new instance of the roleBinding resource.
func NewRoleBinding() resource.Resource {
	return &roleBinding{}
}

func (r *roleBinding) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_binding"
}

func (r *roleBinding) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier of the role binding.",
				MarkdownDescription: "The unique identifier of the role binding.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Description:         "The name of the role this binding applies to.",
				MarkdownDescription: "The name of the role this binding applies to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"groups": schema.ListAttribute{
				Description:         "The groups this role binding applies to.",
				MarkdownDescription: "The groups this role binding applies to.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					validators.GFFieldList(roleBindingValidator, "groups"),
				},
			},
			"users": schema.ListAttribute{
				Description:         "The users this role binding applies to.",
				MarkdownDescription: "The users this role binding applies to.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					validators.GFFieldList(roleBindingValidator, "users"),
				},
			},
		},
	}
}

func (r *roleBinding) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleBinding) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.RBACV1().RoleBindings().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Role Binding", fmt.Sprintf("Could not create Role Binding: %s", err))
		return
	}

	plan = newRoleBindingModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleBinding) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.RBACV1().RoleBindings().Get(ctx, state.ID.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Role Binding",
				fmt.Sprintf("Could not read Role Binding: %s", err),
			)
		}
		return
	}

	updatedState := newRoleBindingModel(obj)
	resp.Diagnostics.Append(normalize.Model(ctx, &updatedState, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *roleBinding) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldObj := state.ToObject()
	newObj := plan.ToObject()

	pb, err := patch.Create(oldObj, newObj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Role Binding Patch",
			fmt.Sprintf("Could not create patch for Role Binding: %v", err),
		)
		return
	}

	patchedObj, err := r.clientSet.RBACV1().RoleBindings().Patch(ctx, oldObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Role Binding",
			fmt.Sprintf("Could not patch Role Binding: %v", err),
		)
		return
	}

	plan = newRoleBindingModel(patchedObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleBinding) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.RBACV1().RoleBindings().Delete(ctx, state.ID.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Role Binding",
			fmt.Sprintf("Could not delete Role Binding: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.RBACV1().RoleBindings(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Role Binding Deletion",
			fmt.Sprintf("Error occurred while waiting for Role Binding to be deleted: %v", err),
		)
		return
	}
}

func (r *roleBinding) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
