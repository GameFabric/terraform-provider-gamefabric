package storage

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	volumereg "github.com/gamefabric/gf-core/pkg/apiserver/registry/storage/volume"
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
	_ resource.Resource                = &volume{}
	_ resource.ResourceWithConfigure   = &volume{}
	_ resource.ResourceWithImportState = &volume{}
)

var volumeValidator = validators.NewGameFabricValidator[*storagev1beta1.Volume, volumeModel](func() validators.StoreValidator {
	storage, _ := volumereg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

type volume struct {
	clientSet clientset.Interface
}

// NewVolume returns a new instance of the volume resource.
func NewVolume() resource.Resource {
	return &volume{}
}

// Metadata defines the resource type name.
func (r *volume) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// Schema defines the schema for this resource.
func (r *volume) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"volume_store": schema.StringAttribute{
				Description:         "The volume store used to store the volume in.",
				MarkdownDescription: "The volume store used to store the volume in.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Immutable.
				},
			},
			"capacity": schema.StringAttribute{
				Description:         "The size of the volume.",
				MarkdownDescription: "The size of the volume.",
				Required:            true,
				Validators: []validator.String{
					validators.GFFieldString(volumeValidator, "spec.capacity"),
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *volume) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volume) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.StorageV1Beta1().Volumes(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Volume",
			fmt.Sprintf("Could not create Volume: %v", err),
		)
		return
	}

	plan = newVolumeModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volume) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.StorageV1Beta1().Volumes(state.Environment.ValueString()).Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Volume",
				fmt.Sprintf("Could not read Volume %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newVolumeModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volume) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state volumeModel
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

	// Check that capacity is not decreased.
	if oldObj.Spec.Capacity.Cmp(newObj.Spec.Capacity) > 0 {
		resp.Diagnostics.AddError(
			"Capacity cannot be decreased",
			fmt.Sprintf("The capacity of the volume %q cannot be decreased", newObj.Name),
		)
		return
	}

	pb, err := patch.Create(oldObj, newObj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Volume Patch",
			fmt.Sprintf("Could not create patch for Volume: %v", err),
		)
		return
	}

	if _, err = r.clientSet.StorageV1Beta1().Volumes(newObj.Environment).Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Volume",
			fmt.Sprintf("Could not patch Volume: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volume) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.StorageV1Beta1().Volumes(state.Environment.ValueString()).Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Volume",
			fmt.Sprintf("Could not delete Volume: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.StorageV1Beta1().Volumes(state.Environment.ValueString()), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Volume Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Volume: %v", err),
		)
		return
	}
}

func (r *volume) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(req.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), env)...)
}
