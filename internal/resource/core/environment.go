package core

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
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
	_ resource.Resource                = &environment{}
	_ resource.ResourceWithConfigure   = &environment{}
	_ resource.ResourceWithImportState = &environment{}
)

type environment struct {
	clientSet clientset.Interface
}

// NewEnvironment returns a new environment resource.
func NewEnvironment() resource.Resource {
	return &environment{}
}

// Metadata sets the resource type name.
func (r *environment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// Schema defines the schema for this data source.
func (r *environment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Environment resource represents a logical grouping and isolation of game servers, configurations, and other related resources within GameFabric.",
		MarkdownDescription: "Environment resource represents a logical grouping and isolation of game servers, configurations, and other related resources within GameFabric.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "The unique environment object name. The name must be between 1 and 4 characters long and only contain alphanumeric lowercase characters without any special characters.",
				MarkdownDescription: "The unique environment object name. The name must be between 1 and 4 characters long and only contain alphanumeric lowercase characters without any special characters.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is the name of the environment shown in GameFabric.",
				MarkdownDescription: "DisplayName is the name of the environment shown in GameFabric.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Description:         "Description is the description of the environment.",
				MarkdownDescription: "Description is the description of the environment.",
				Optional:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *environment) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	_, err := r.clientSet.CoreV1().Environments().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Environment",
			fmt.Sprintf("Could not create Environment: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(obj.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.CoreV1().Environments().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Environment",
				fmt.Sprintf("Could not read Environment %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newEnvironmentModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state environmentModel
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
			"Error Creating Environment Patch",
			fmt.Sprintf("Could not create patch for Environment: %v", err),
		)
		return
	}

	if _, err = r.clientSet.CoreV1().Environments().Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Environment",
			fmt.Sprintf("Could not patch for Environment: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(newObj.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.CoreV1().Environments().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Environment",
			fmt.Sprintf("Could not delete for Environment: %v", err),
		)
		return
	}

	err = wait.PollUntilNotFound(ctx, r.clientSet.CoreV1().Environments(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Environment Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Environment: %v", err),
		)
	}
}

func (r *environment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
