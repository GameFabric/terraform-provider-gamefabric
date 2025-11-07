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
	policyreg "github.com/gamefabric/gf-core/pkg/apiserver/registry/storage/volumestoreretentionpolicy"
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
	_ resource.Resource                = &volumeStoreRetentionPolicy{}
	_ resource.ResourceWithConfigure   = &volumeStoreRetentionPolicy{}
	_ resource.ResourceWithImportState = &volumeStoreRetentionPolicy{}
)

var retPolicyValidator = validators.NewGameFabricValidator[*storagev1beta1.VolumeStoreRetentionPolicy, volumeStoreRetentionPolicyModel](
	func() validators.StoreValidator {
		storage, _ := policyreg.New(generic.StoreOptions{Config: generic.Config{StorageFactory: registrytest.FakeStorageFactory{}}})
		return storage.Store.Strategy
	},
)

type volumeStoreRetentionPolicy struct {
	clientSet clientset.Interface
}

// NewVolumeStoreRetentionPolicy returns a new instance of the volume store retention policy resource.
func NewVolumeStoreRetentionPolicy() resource.Resource {
	return &volumeStoreRetentionPolicy{}
}

// Metadata defines the resource type name.
func (r *volumeStoreRetentionPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumestore_retention_policy"
}

// Schema defines the schema for this data source.
func (r *volumeStoreRetentionPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
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
			"volume_store": schema.StringAttribute{
				Description:         "The name of the volume to find snapshots on.",
				MarkdownDescription: "The name of the volume to find snapshots on.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"offline_snapshot": schema.SingleNestedAttribute{
				Description:         "The retention policy for offline snapshots.",
				MarkdownDescription: "The retention policy for offline snapshots.",
				Optional:            true,
				Computed:            true,
				Default:             (*volumeStoreRetentionPolicySnapshot)(nil).Default(),
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						Description:         "Count is the minimum number of offline snapshots to keep.",
						MarkdownDescription: "Count is the minimum number of offline snapshots to keep.",
						Optional:            true,
						Validators: []validator.Int64{
							validators.GFFieldInt64(retPolicyValidator, "spec.snapshots.offlineCount"),
						},
					},
					"days": schema.Int64Attribute{
						Description:         "Days is the minimum number of days to keep offline snapshots for.",
						MarkdownDescription: "Days is the minimum number of days to keep offline snapshots for.",
						Optional:            true,
						Validators: []validator.Int64{
							validators.GFFieldInt64(retPolicyValidator, "spec.snapshots.offlineDays"),
						},
					},
				},
			},
			"online_snapshot": schema.SingleNestedAttribute{
				Description:         "The retention policy for online snapshots.",
				MarkdownDescription: "The retention policy for online snapshots.",
				Optional:            true,
				Computed:            true,
				Default:             (*volumeStoreRetentionPolicySnapshot)(nil).Default(),
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						Description:         "Count is the minimum number of online snapshots to keep.",
						MarkdownDescription: "Count is the minimum number of online snapshots to keep.",
						Optional:            true,
						Validators: []validator.Int64{
							validators.GFFieldInt64(retPolicyValidator, "spec.snapshots.onlineCount"),
						},
					},
					"days": schema.Int64Attribute{
						Description:         "Days is the minimum number of days to keep online snapshots for.",
						MarkdownDescription: "Days is the minimum number of days to keep online snapshots for.",
						Optional:            true,
						Validators: []validator.Int64{
							validators.GFFieldInt64(retPolicyValidator, "spec.snapshots.onlineDays"),
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *volumeStoreRetentionPolicy) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumeStoreRetentionPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeStoreRetentionPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.StorageV1Beta1().VolumeStoreRetentionPolicies(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating VolumeStoreRetentionPolicy",
			fmt.Sprintf("Could not create VolumeStoreRetentionPolicy: %v", err),
		)
		return
	}

	plan = newVolumeStoreRetentionPolicyModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeStoreRetentionPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeStoreRetentionPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.StorageV1Beta1().
		VolumeStoreRetentionPolicies(state.Environment.ValueString()).
		Get(ctx, state.VolumeStore.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading VolumeStoreRetentionPolicy",
				fmt.Sprintf("Could not read VolumeStoreRetentionPolicy for volume store %q: %v", state.VolumeStore.ValueString(), err),
			)
		}
		return
	}

	state = newVolumeStoreRetentionPolicyModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeStoreRetentionPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state volumeStoreRetentionPolicyModel
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
			"Error Creating VolumeStoreRetentionPolicy Patch",
			fmt.Sprintf("Could not create patch for VolumeStoreRetentionPolicy: %v", err),
		)
		return
	}

	_, err = r.clientSet.StorageV1Beta1().
		VolumeStoreRetentionPolicies(newObj.Environment).
		Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Patching VolumeStoreRetentionPolicy",
			fmt.Sprintf("Could not patch for VolumeStoreRetentionPolicy: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeStoreRetentionPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeStoreRetentionPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.StorageV1Beta1().
		VolumeStoreRetentionPolicies(state.Environment.ValueString()).
		Delete(ctx, state.VolumeStore.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting VolumeStoreRetentionPolicy",
			fmt.Sprintf("Could not delete VolumeStoreRetentionPolicy: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.StorageV1Beta1().VolumeStoreRetentionPolicies(state.Environment.ValueString()), state.VolumeStore.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for VolumeStoreRetentionPolicy Deletion",
			fmt.Sprintf("Timed out waiting for deletion of VolumeStoreRetentionPolicy: %v", err),
		)
		return
	}
}

func (r *volumeStoreRetentionPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(req.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_store"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), env)...)
}
