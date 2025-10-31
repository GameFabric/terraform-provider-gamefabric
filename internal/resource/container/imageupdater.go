package container

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &imageUpdater{}
	_ resource.ResourceWithConfigure = &imageUpdater{}
)

type imageUpdater struct {
	clientSet clientset.Interface
}

// NewImageUpdater returns a new instance of the image updater resource.
func NewImageUpdater() resource.Resource {
	return &imageUpdater{}
}

// Metadata defines the resource type name.
func (r *imageUpdater) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_imageupdater"
}

// Schema defines the schema for this data source.
func (r *imageUpdater) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
			},
			"branch": schema.StringAttribute{
				Description:         "The branch that contains the image.",
				MarkdownDescription: "The branch that contains the image.",
				Required:            true,
			},
			"image": schema.StringAttribute{
				Description:         "The name of the image to update.",
				MarkdownDescription: "The name of the image to update.",
				Required:            true,
			},
			"target": schema.SingleNestedAttribute{
				Description:         "Target configuration specifying the resource to update.",
				MarkdownDescription: "Target configuration specifying the resource to update.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description:         "The type of the target resource.",
						MarkdownDescription: "The type of the target resource.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								ImageUpdaterTargetTypeArmadaSet, ImageUpdaterTargetTypeArmada, ImageUpdaterTargetTypeFormation, ImageUpdaterTargetTypeVessel,
							),
						},
					},
					"name": schema.StringAttribute{
						Description:         "The name of the target resource.",
						MarkdownDescription: "The name of the target resource.",
						Required:            true,
						Validators: []validator.String{
							validators.NameValidator{},
						},
					},
					"environment": schema.StringAttribute{
						Description:         "The environment in which the target resource operates.",
						MarkdownDescription: "The environment in which the target resource operates.",
						Required:            true,
						Validators: []validator.String{
							validators.EnvironmentValidator{},
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *imageUpdater) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageUpdater) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageUpdaterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject(uuid.New().String())
	outObj, err := r.clientSet.ContainerV1().ImageUpdaters(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Image Updater",
			fmt.Sprintf("Could not create ImageUpdater: %v", err),
		)
		return
	}

	plan = newImageUpdaterModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imageUpdater) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageUpdaterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(state.ID.ValueString())
	outObj, err := r.clientSet.ContainerV1().ImageUpdaters(env).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Image Updater",
				fmt.Sprintf("Could not read ImageUpdater for branch %q and image %q: %v", state.Branch.ValueString(), state.Image.ValueString(), err),
			)
		}
		return
	}

	state = newImageUpdaterModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageUpdater) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state imageUpdaterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(state.ID.ValueString())

	oldObj := state.ToObject(name)
	newObj := plan.ToObject(name)

	pb, err := patch.Create(oldObj, newObj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Image Updater Patch",
			fmt.Sprintf("Could not create patch for ImageUpdater: %v", err),
		)
		return
	}

	if _, err = r.clientSet.ContainerV1().ImageUpdaters(env).Patch(ctx, name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Image Updater",
			fmt.Sprintf("Could not patch for ImageUpdater: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imageUpdater) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageUpdaterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(state.ID.ValueString())
	err := r.clientSet.ContainerV1().ImageUpdaters(env).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Image Updater",
			fmt.Sprintf("Could not delete for ImageUpdater: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.ContainerV1().ImageUpdaters(env), name); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Image Updater Deletion",
			fmt.Sprintf("Timed out waiting for deletion of ImageUpdater: %v", err),
		)
		return
	}
}
