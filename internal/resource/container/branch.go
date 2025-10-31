package container

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
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
	_ resource.Resource                = &branch{}
	_ resource.ResourceWithConfigure   = &branch{}
	_ resource.ResourceWithImportState = &branch{}
)

// branch is the branch resource.
type branch struct {
	clientSet clientset.Interface
}

// NewBranch creates a new branch resource.
func NewBranch() resource.Resource {
	return &branch{}
}

// Metadata returns the resource metadata.
func (r *branch) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch"
}

// Schema defines the schema for the branch resource.
func (r *branch) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
				MarkdownDescription: "The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description:         "The user-friendly name of the branch.",
				MarkdownDescription: "The user-friendly name of the branch.",
				Optional:            true,
			},
			"labels": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.LabelsValidator{},
				},
			},
			"annotations": schema.MapAttribute{
				Description:         "Annotations is an unstructured map of keys and values stored on an object.",
				MarkdownDescription: "Annotations is an unstructured map of keys and values stored on an object.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.AnnotationsValidator{},
				},
			},
			"description": schema.StringAttribute{
				Description:         "Description is the optional description of the branch.",
				MarkdownDescription: "Description is the optional description of the branch.",
				Optional:            true,
			},
			"retention_policy_rules": schema.ListNestedAttribute{
				Description:         "RetentionPolicyRules are the rules that define how images are retained.",
				MarkdownDescription: "RetentionPolicyRules are the rules that define how images are retained.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the image retention policy.",
							MarkdownDescription: "Name is the name of the image retention policy.",
							Required:            true,
						},
						"image_regex": schema.StringAttribute{
							Description:         "ImageRegex is the optional regex selector for images that this policy applies to.",
							MarkdownDescription: "ImageRegex is the optional regex selector for images that this policy applies to.",
							Optional:            true,
						},
						"tag_regex": schema.StringAttribute{
							Description:         "TagRegex is the optional regex selector for tags that this policy applies to.",
							MarkdownDescription: "TagRegex is the optional regex selector for tags that this policy applies to.",
							Optional:            true,
						},
						"keep_count": schema.Int64Attribute{
							Description:         "KeepCount is the minimum number of tags to keep per image.",
							MarkdownDescription: "KeepCount is the minimum number of tags to keep per image.",
							Optional:            true,
						},
						"keep_days": schema.Int64Attribute{
							Description:         "KeepDays is the minimum number of days an image tag must be kept for.",
							MarkdownDescription: "KeepDays is the minimum number of days an image tag must be kept for.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *branch) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	procCtx, ok := req.ProviderData.(*provcontext.Context)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *provcontext.Context, got: %T", req.ProviderData),
		)
		return
	}

	r.clientSet = procCtx.ClientSet
}

// Create creates the resource and sets the initial Terraform state on success.
func (r *branch) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan branchModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.ContainerV1().Branches().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Error",
			fmt.Sprintf("Could not create Branch %q: %v", plan.Name.ValueString(), err),
		)
		return
	}
	plan = newBranchModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *branch) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state branchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.ContainerV1().Branches().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Read Error",
				fmt.Sprintf("Could not read Branch %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newBranchModel(obj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *branch) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state branchModel
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
			"Creating Branch Patch",
			fmt.Sprintf("Could not create patch for Branch %q: %v", state.Name.ValueString(), err),
		)
		return
	}

	outObj, err := r.clientSet.ContainerV1().Branches().Patch(ctx, state.Name.ValueString(), rest.MergePatchType, pb, metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Patching Branch",
			fmt.Sprintf("Could not patch Branch %q: %v", state.Name.ValueString(), err),
		)
		return
	}

	plan = newBranchModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *branch) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state branchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.clientSet.ContainerV1().Branches().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Branch",
			fmt.Sprintf("Could not delete Branch %q: %v", state.Name.ValueString(), err),
		)
		return
	}

	if err := wait.PollUntilNotFound(ctx, r.clientSet.ContainerV1().Branches(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Branch Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Branch: %v", err),
		)
	}
}

func (r *branch) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
