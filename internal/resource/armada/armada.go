package armada

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	armadareg "github.com/gamefabric/gf-core/pkg/apiserver/registry/armada/armada"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &armada{}
	_ resource.ResourceWithConfigure   = &armada{}
	_ resource.ResourceWithImportState = &armada{}
)

var armadaValidator = validators.NewGameFabricValidator[*armadav1.Armada, armadaModel](func() validators.StoreValidator {
	storage, _ := armadareg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

type armada struct {
	clientSet clientset.Interface
}

// NewArmada returns a new instance of the Armada resource.
func NewArmada() resource.Resource {
	return &armada{}
}

// Metadata defines the resource type name.
func (r *armada) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_armada"
}

// Schema defines the schema for this data source.
func (r *armada) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) { //nolint:maintidx // Keep schema in one place.
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 49 characters.",
				MarkdownDescription: "The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 49 characters.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					validators.GFFieldString(armadaValidator, "metadata.name"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the resource belongs to.",
				MarkdownDescription: "The name of the environment the resource belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description:         "Description is the optional description of the armada.",
				MarkdownDescription: "Description is the optional description of the armada.",
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
			"autoscaling": schema.SingleNestedAttribute{
				Description:         "Autoscaling configuration for the game servers.",
				MarkdownDescription: "Autoscaling configuration for the game servers.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"fixed_interval_seconds": schema.Int32Attribute{
						Description:         "Defines how often the auto-scaler re-evaluates the number of game servers.",
						MarkdownDescription: "Defines how often the auto-scaler re-evaluates the number of game servers.",
						Optional:            true,
						Validators: []validator.Int32{
							int32validator.AtLeast(0), // 0 is like omitempty.
							validators.GFFieldInt32(armadaValidator, "spec.autoscaling.fixedInterval.seconds"),
						},
					},
				},
			},
			"region": schema.StringAttribute{
				Description:         "Region defines the region the game servers are distributed to.",
				MarkdownDescription: "Region defines the region the game servers are distributed to.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					validators.GFFieldString(armadaValidator, "spec.region"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Region is immutable.
				},
			},
			"replicas": schema.ListNestedAttribute{
				Description:         "A replicas specifies the distribution of game servers across the available types of capacity in the selected region type.",
				MarkdownDescription: "A replicas specifies the distribution of game servers across the available types of capacity in the selected region type.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"region_type": schema.StringAttribute{
							Description:         "RegionType is the name of the region type.",
							MarkdownDescription: "RegionType is the name of the region type.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
								validators.GFFieldString(armadaValidator, "spec.distribution[?].name"),
							},
						},
						"min_replicas": schema.Int32Attribute{
							Description:         "MinReplicas is the minimum number of replicas in the region type.",
							MarkdownDescription: "MinReplicas is the minimum number of replicas in the region type.",
							Required:            true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
								int32validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("buffer_size")),
								validators.GFFieldInt32(armadaValidator, "spec.distribution[?].minReplicas"),
							},
						},
						"max_replicas": schema.Int32Attribute{
							Description:         "MaxReplicas is the maximum number of replicas in the region type.",
							MarkdownDescription: "MaxReplicas is the maximum number of replicas in the region type.",
							Required:            true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
								validators.GFFieldInt32(armadaValidator, "spec.distribution[?].maxReplicas"),
							},
						},
						"buffer_size": schema.Int32Attribute{
							Description:         "BufferSize is the number of replicas to have ready all the time.",
							MarkdownDescription: "BufferSize is the number of replicas to have ready all the time.",
							Required:            true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
								validators.GFFieldInt32(armadaValidator, "spec.distribution[?].bufferSize"),
							},
						},
					},
				},
			},
			"gameserver_labels": schema.MapAttribute{
				Description:         "Labels for the game server pods.",
				MarkdownDescription: "Labels for the game server pods.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.LabelsValidator{},
				},
			},
			"gameserver_annotations": schema.MapAttribute{
				Description:         "Annotations for the game server pods.",
				MarkdownDescription: "Annotations for the game server pods.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.AnnotationsValidator{},
				},
			},
			"containers": schema.ListNestedAttribute{
				Description:         "Containers is a list of containers belonging to the game server.",
				MarkdownDescription: "Containers is a list of containers belonging to the game server.",
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: mps.ContainersAttributes(armadaValidator, "spec.template.spec.containers[?]"),
			},
			"health_checks": schema.SingleNestedAttribute{
				Description:         "HealthChecks is the health checking configuration for Agones game servers.",
				MarkdownDescription: "HealthChecks is the health checking configuration for Agones game servers.",
				Optional:            true,
				Attributes:          mps.HealthCheckAttributes(armadaValidator, "spec.template.spec.health"),
			},
			"termination_configuration": schema.SingleNestedAttribute{
				Description:         "TerminationConfiguration defines the termination grace period for game servers.",
				MarkdownDescription: "TerminationConfiguration defines the termination grace period for game servers.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"grace_period_seconds": schema.Int64Attribute{
						Description:         "GracePeriodSeconds is the duration in seconds the game server needs to terminate gracefully.",
						MarkdownDescription: "GracePeriodSeconds is the duration in seconds the game server needs to terminate gracefully.",
						Optional:            true,
						Validators: []validator.Int64{
							validators.GFFieldInt64(armadaValidator, "spec.template.spec.terminationGracePeriodSeconds"),
						},
					},
				},
			},
			"strategy": schema.SingleNestedAttribute{
				Description:         "Strategy defines the rollout strategy for updating game servers. The default is RollingUpdate.",
				MarkdownDescription: "Strategy defines the rollout strategy for updating game servers. The default is RollingUpdate.",
				Optional:            true,
				Validators: []validator.Object{
					validators.GFFieldObject(armadaValidator, "spec.template.spec.strategy"),
				},
				Attributes: map[string]schema.Attribute{
					"rolling_update": schema.SingleNestedAttribute{
						Description:         "RollingUpdate defines the rolling update strategy, which gradually replaces game servers with new ones.",
						MarkdownDescription: "RollingUpdate defines the rolling update strategy, which gradually replaces game servers with new ones.",
						Optional:            true,
						Validators: []validator.Object{
							objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("recreate")),
							validators.GFFieldObject(armadaValidator, "spec.template.spec.strategy.rollingUpdate"),
						},
						Attributes: map[string]schema.Attribute{
							"max_surge": schema.StringAttribute{
								Description:         "MaxSurge is the maximum number of game servers that can be created over the desired number of game servers during an update.",
								MarkdownDescription: "MaxSurge is the maximum number of game servers that can be created over the desired number of game servers during an update.",
								Optional:            true,
								Validators: []validator.String{
									validators.GFFieldString(armadaValidator, "spec.template.spec.strategy.rollingUpdate.maxSurge"),
								},
							},
							"max_unavailable": schema.StringAttribute{
								Description:         "MaxUnavailable is the maximum number of game servers that can be unavailable during an update.",
								MarkdownDescription: "MaxUnavailable is the maximum number of game servers that can be unavailable during an update.",
								Optional:            true,
								Validators: []validator.String{
									validators.GFFieldString(armadaValidator, "spec.template.spec.strategy.rollingUpdate.maxUnavailable"),
								},
							},
						},
					},
					"recreate": schema.SingleNestedAttribute{
						Description:         "Recreate defines the recreate strategy which will recreate all unallocated game servers at once. This should only be used for development workloads or where downtime is acceptable.",
						MarkdownDescription: "Recreate defines the recreate strategy which will recreate all unallocated game servers at once on updates. This should only be used for development workloads or where downtime is acceptable.",
						Optional:            true,
						Validators: []validator.Object{
							objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("rolling_update")),
						},
					},
				},
			},
			"volumes": schema.ListNestedAttribute{
				Description:         "Volumes is a list of volumes that can be mounted by containers belonging to the game server.",
				MarkdownDescription: "Volumes is a list of volumes that can be mounted by containers belonging to the game server.",
				Optional:            true,
				Validators: []validator.List{
					validators.GFFieldList(armadaValidator, "spec.template.spec.volumes"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the volume.",
							MarkdownDescription: "Name is the name of the volume.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
								validators.GFFieldString(armadaValidator, "spec.template.spec.volumes[?].name"),
							},
						},
						"empty_dir": schema.SingleNestedAttribute{
							Description:         "EmptyDir represents a volume that is initially empty.",
							MarkdownDescription: "EmptyDir represents a volume that is initially empty.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"size_limit": schema.StringAttribute{
									Description:         "SizeLimit is the total amount of local storage required for this EmptyDir volume.",
									MarkdownDescription: "SizeLimit is the total amount of local storage required for this EmptyDir volume.",
									Optional:            true,
									Validators: []validator.String{
										validators.QuantityValidator{},
										validators.GFFieldString(armadaValidator, "spec.template.spec.volumes[?].sizeLimit"),
									},
								},
							},
						},
					},
				},
			},
			"gateway_policies": schema.ListAttribute{
				Description:         "GatewayPolicies is a list of gateway policies to apply to the Armada.",
				MarkdownDescription: "GatewayPolicies is a list of gateway policies to apply to the Armada.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validators.NameValidator{},
						validators.GFFieldString(armadaValidator, "spec.template.spec.gatewayPolicies[?]"),
					),
				},
			},
			"profiling_enabled": schema.BoolAttribute{
				Description:         "ProfilingEnabled indicates whether profiling is enabled for the Armada.",
				MarkdownDescription: "ProfilingEnabled indicates whether profiling is enabled for the Armada.",
				Optional:            true,
			},
			"image_updater_target": schema.SingleNestedAttribute{
				Description:         "ImageUpdaterTarget is the reference that an image updater can target to match the Armada.",
				MarkdownDescription: "ImageUpdaterTarget is the reference that an image updater can target to match the Armada.",
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description:         "Type is the type of the image updater target.",
						MarkdownDescription: "Type is the type of the image updater target.",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						Description:         "Name is the name of the image updater target.",
						MarkdownDescription: "Name is the name of the image updater target.",
						Computed:            true,
					},
					"environment": schema.StringAttribute{
						Description:         "Environment is the environment of the image updater target.",
						MarkdownDescription: "Environment is the environment of the image updater target.",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *armada) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *armada) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan armadaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.ArmadaV1().Armadas(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Armada",
			fmt.Sprintf("Could not create Armada: %v", err),
		)
		return
	}

	state := newArmadaModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.Plan)...)
	// Preserve quantity values from plan when semantically equal (e.g., "1000m" vs "1")
	resp.Diagnostics.Append(mps.PreserveContainerQuantities(ctx, &state, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *armada) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState armadaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.ArmadaV1().Armadas(oldState.Environment.ValueString()).Get(ctx, oldState.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Armada",
				fmt.Sprintf("Could not read Armada %q: %v", oldState.Name.ValueString(), err),
			)
		}
		return
	}

	newState := newArmadaModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &newState, req.State)...)
	// Preserve quantity values from old state when semantically equal (e.g., "1000m" vs "1")
	resp.Diagnostics.Append(mps.PreserveContainerQuantities(ctx, &newState, &oldState)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *armada) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state armadaModel
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
			"Error Creating Armada Patch",
			fmt.Sprintf("Could not create patch for Armada: %v", err),
		)
		return
	}

	if _, err = r.clientSet.ArmadaV1().Armadas(newObj.Environment).Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Armada",
			fmt.Sprintf("Could not patch Armada: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	plan.ImageUpdaterTarget = container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeArmada, oldObj.Name, oldObj.Environment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *armada) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state armadaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.ArmadaV1().Armadas(state.Environment.ValueString()).Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Armada",
			fmt.Sprintf("Could not delete for Armada: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.ArmadaV1().Armadas(state.Environment.ValueString()), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Armada Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Armada: %v", err),
		)
		return
	}
}

func (r *armada) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(req.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), env)...)
}
