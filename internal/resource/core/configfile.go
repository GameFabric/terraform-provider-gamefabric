package core

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/cache"
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
	_ resource.Resource                = &configFile{}
	_ resource.ResourceWithConfigure   = &configFile{}
	_ resource.ResourceWithImportState = &configFile{}
)

type configFile struct {
	clientSet clientset.Interface
}

// NewConfigFile returns a new instance of the config file resource.
func NewConfigFile() resource.Resource {
	return &configFile{}
}

// Metadata defines the resource type name.
func (r *configFile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configfile"
}

// Schema defines the schema for this resource.
func (r *configFile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope.",
				MarkdownDescription: "The unique object name within its scope.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the object belongs to.",
				MarkdownDescription: "The name of the environment the object belongs to.",
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
			"description": schema.StringAttribute{
				Description:         "Description is the optional description of the config file.",
				MarkdownDescription: "Description is the optional description of the config file.",
				Optional:            true,
			},
			"data": schema.StringAttribute{
				Description:         "The content of the config file.",
				MarkdownDescription: "The content of the config file.",
				Required:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *configFile) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *configFile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	_, err := r.clientSet.CoreV1().ConfigFiles(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Config File",
			fmt.Sprintf("Could not create ConfigFile: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *configFile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.CoreV1().ConfigFiles(state.Environment.ValueString()).Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Config File",
				fmt.Sprintf("Could not read ConfigFile %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newConfigModel(outObj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *configFile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state configFileModel
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
			"Error Creating Config File Patch",
			fmt.Sprintf("Could not create patch for ConfigFile: %v", err),
		)
		return
	}

	if _, err = r.clientSet.CoreV1().ConfigFiles(newObj.Environment).Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Config File",
			fmt.Sprintf("Could not patch ConfigFile: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *configFile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state configFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.CoreV1().ConfigFiles(state.Environment.ValueString()).Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Config File",
			fmt.Sprintf("Could not delete ConfigFile: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.CoreV1().ConfigFiles(state.Environment.ValueString()), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Config File Deletion",
			fmt.Sprintf("Timed out waiting for deletion of ConfigFile: %v", err),
		)
		return
	}
}

func (r *configFile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(req.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), env)...)
}
