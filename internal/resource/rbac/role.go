package rbac

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &role{}
	_ resource.ResourceWithConfigure   = &role{}
	_ resource.ResourceWithImportState = &role{}
)

type role struct {
	clientSet clientset.Interface
}

// NewRole returns a new instance of the role resource.
func NewRole() resource.Resource {
	return &role{}
}

func (r *role) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *role) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Unique ID of the role.",
				MarkdownDescription: "Unique ID of the role.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "The unique name of the role.",
				MarkdownDescription: "The unique name of the role.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rules": schema.ListNestedAttribute{
				Description:         "List of rules that will be applied to the role.",
				MarkdownDescription: "List of rules that will be applied to the role.",
				Required:            true,
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"verbs": schema.ListAttribute{
							Description:         "List of actions that can be performed on the resources. Use `*` to match all verbs.",
							MarkdownDescription: "List of actions that can be performed on the resources. Use `*` to match all verbs.",
							ElementType:         types.StringType,
							Required:            true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.OneOf("get", "watch", "list", "post", "put", "patch", "delete", "*"),
								),
							},
						},
						"api_groups": schema.ListAttribute{
							Description:         "List of API groups that the rule applies to. Use `*` to match all API groups or choose any from: `armada`, `authentication`, `billing`, `container`, `core`, `formation`, `protection`, `rbac` and `storage`.",
							MarkdownDescription: "List of API groups that the rule applies to. Use `*` to match all API groups or choose any from: `armada`, `authentication`, `billing`, `container`, `core`, `formation`, `protection`, `rbac` and `storage`.",
							ElementType:         types.StringType,
							Required:            true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.Any(
										stringvalidator.RegexMatches(regexp.MustCompile("^[a-z]+$"), "must be a valid API group format"),
										stringvalidator.OneOfCaseInsensitive("*"),
									),
								),
							},
						},
						"environments": schema.ListAttribute{
							Description:         "List of environments that the rule applies to. Use `*` to match all environments.",
							MarkdownDescription: "List of environments that the rule applies to. Use `*` to match all environments.",
							ElementType:         types.StringType,
							Required:            true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.Any(
										validators.EnvironmentValidator{},
										stringvalidator.OneOfCaseInsensitive("*"),
									),
								),
							},
						},
						"resources": schema.ListAttribute{
							Description:         "List of resource types that the rule applies to. Use `*` to match all resources, use the plural for specific resources, e.g. `armadas`, use the slash for sub-resources, e.g. `images/status`.",
							MarkdownDescription: "List of resource types that the rule applies to. Use '*' to match all resources, use the plural for specific resources, e.g. `armadas`, use the slash for sub-resources, e.g. `images/status`.",
							ElementType:         types.StringType,
							Required:            true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.Any(
										stringvalidator.RegexMatches(regexp.MustCompile("^[a-z]+(/[a-z]+)?$"), "must be a valid resource or sub-resource format"),
										stringvalidator.OneOfCaseInsensitive("*"),
									),
								),
							},
						},
						"scopes": schema.ListAttribute{
							Description:         "List of scopes to restrict the rule to. Optional field to further limit the permissions.",
							MarkdownDescription: "List of scopes to restrict the rule to. Optional field to further limit the permissions.",
							ElementType:         types.StringType,
							Optional:            true,
						},
						"resource_names": schema.ListAttribute{
							Description:         "List of specific resource names that the rule applies to. If specified, the rule only applies to resources with these names.",
							MarkdownDescription: "List of specific resource names that the rule applies to. If specified, the rule only applies to resources with these names.",
							ElementType:         types.StringType,
							Optional:            true,
						},
					},
				},
			},
			"labels": schema.MapAttribute{
				Description:         "A map of labels to assign to the role.",
				MarkdownDescription: "A map of labels to assign to the role.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.LabelsValidator{},
				},
			},
			"annotations": schema.MapAttribute{
				Description:         "A map of annotations to assign to the role.",
				MarkdownDescription: "A map of annotations to assign to the role.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.AnnotationsValidator{},
				},
			},
		},
	}
}

func (r *role) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *role) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.RBACV1().Roles().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Role", fmt.Sprintf("Could not create Role: %v", err))
		return
	}

	plan = newRoleModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.RBACV1().Roles().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Role",
				fmt.Sprintf("Could not read Role: %v", err),
			)
		}
		return
	}

	state = newRoleModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *role) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleModel
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
			"Error Creating Role Patch",
			fmt.Sprintf("Could not create patch for Role: %v", err),
		)
		return
	}

	if _, err = r.clientSet.RBACV1().Roles().Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Role",
			fmt.Sprintf("Could not patch Role: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(newObj.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.RBACV1().Roles().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Role",
			fmt.Sprintf("Could not delete Role: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.RBACV1().Roles(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Role Deletion",
			fmt.Sprintf("Error occurred while waiting for Role to be deleted: %v", err),
		)
		return
	}
}

func (r *role) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
