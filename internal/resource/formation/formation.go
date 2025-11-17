package formation

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	formationreg "github.com/gamefabric/gf-core/pkg/apiserver/registry/formation/formation"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &formation{}
	_ resource.ResourceWithConfigure   = &formation{}
	_ resource.ResourceWithImportState = &formation{}
)

var formationValidator = validators.NewGameFabricValidator[*formationv1.Formation, formationModel](func() validators.StoreValidator {
	storage, _ := formationreg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

type formation struct {
	clientSet clientset.Interface
}

// NewFormation returns a new instance of the Formation resource.
func NewFormation() resource.Resource {
	return &formation{}
}

// Metadata defines the resource type name.
func (r *formation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_formation"
}

// Schema defines the schema for this data source.
func (r *formation) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) { //nolint:maintidx // Keep schema in one place.
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
			"description": schema.StringAttribute{
				Description:         "Description is the optional description of the Formation.",
				MarkdownDescription: "Description is the optional description of the Formation.",
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
			"volume_templates": schema.ListNestedAttribute{
				Description:         "VolumeTemplates are the templates for volumes that can be mounted by containers belonging to the game server.",
				MarkdownDescription: "VolumeTemplates are the templates for volumes that can be mounted by containers belonging to the game server.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the volume as it will be referenced.",
							MarkdownDescription: "Name is the name of the volume as it will be referenced.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"reclaim_policy": schema.StringAttribute{
							Description: "The reclaim policy for a volume created by a Formation. Defaults to &#34;Packed&#34;",
							MarkdownDescription: `The reclaim policy for a volume created by a Formation.

**Delete:** (recommended) This means that a backing Kubernetes volume is automatically deleted when the user deletes the persisent volume.

**Retain:** If a user deletes the persistent volume, the backing Kubernetes volume will not be deleted. Instead, it is moved to the Released phase, where all of its data can be manually recovered by Nitrado for an undetermined time.`,
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("Retain", "Delete"),
							},
						},
						"volume_store_name": schema.StringAttribute{
							Description:         "Name is the name of the VolumeStore.",
							MarkdownDescription: "Name is the name of the VolumeStore.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"capacity": schema.StringAttribute{
							Description:         "Capacity is the storage capacity requested for the volume.",
							MarkdownDescription: "Capacity is the storage capacity requested for the volume.",
							Required:            true,
							Validators: []validator.String{
								validators.QuantityValidator{},
							},
						},
					},
				},
			},
			"vessels": schema.ListNestedAttribute{
				Description:         "Vessels is a list of vessels belonging to the game server.",
				MarkdownDescription: "Vessels is a list of vessels belonging to the game server.",
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the vessel.",
							MarkdownDescription: "Name is the name of the vessel.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"region": schema.StringAttribute{
							Description:         "Region defines the region the game servers are distributed to.",
							MarkdownDescription: "Region defines the region the game servers are distributed to.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"description": schema.StringAttribute{
							Description:         "Description is the optional description of the Vessel.",
							MarkdownDescription: "Description is the optional description of the Vessel.",
							Optional:            true,
						},
						"suspend": schema.BoolAttribute{
							Description:         "Suspend indicates whether the Vessel should be suspended.",
							MarkdownDescription: "Suspend indicates whether the Vessel should be suspended.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"override": schema.SingleNestedAttribute{
							Description:         "Override is the optional override configuration for Vessels.",
							MarkdownDescription: "Override is the optional override configuration for Vessels.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"gameserver_labels": schema.MapAttribute{
									Description:         "A map of keys and values that can be used to organize and categorize objects.",
									MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
									Optional:            true,
									ElementType:         types.StringType,
									Validators: []validator.Map{
										validators.LabelsValidator{},
									},
								},
								"containers": schema.ListNestedAttribute{
									Description:         "Containers is a list of containers belonging to the game server.",
									MarkdownDescription: "Containers is a list of containers belonging to the game server.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"command": schema.ListAttribute{
												Description:         "Command is the entrypoint array. This is not executed within a shell.",
												MarkdownDescription: "Command is the entrypoint array. This is not executed within a shell.",
												Optional:            true,
												ElementType:         types.StringType,
											},
											"args": schema.ListAttribute{
												Description:         "Args are arguments to the entrypoint.",
												MarkdownDescription: "Args are arguments to the entrypoint.",
												Optional:            true,
												ElementType:         types.StringType,
											},
											"envs": schema.ListNestedAttribute{
												Description:         "Envs is a list of environment variables to set on all containers in this Armada.",
												MarkdownDescription: "Envs is a list of environment variables to set on all containers in this Armada.",
												Optional:            true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: core.EnvVarAttributes(formationValidator, "spec.vessels[?].override.containers[?].env[?]"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"gameserver_labels": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.LabelsValidator{},
				},
			},
			"gameserver_annotations": schema.MapAttribute{
				Description:         "Annotations is an unstructured map of keys and values stored on an object.",
				MarkdownDescription: "Annotations is an unstructured map of keys and values stored on an object.",
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
				NestedObject: mps.ContainersAttributes(formationValidator, "spec.template.spec.containers[?]"),
			},
			"health_checks": schema.SingleNestedAttribute{
				Description:         "HealthChecks is the health checking configuration for Agones game servers.",
				MarkdownDescription: "HealthChecks is the health checking configuration for Agones game servers.",
				Optional:            true,
				Attributes:          mps.HealthCheckAttributes(formationValidator, "spec.template.spec.health"),
			},
			"termination_configuration": schema.SingleNestedAttribute{
				Description:         "TerminationConfiguration defines the termination grace period for game servers.",
				MarkdownDescription: "TerminationConfiguration defines the termination grace period for game servers.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"grace_period_seconds": schema.Int64Attribute{
						Description:         "The duration in seconds the game server needs to terminate gracefully.",
						MarkdownDescription: "The duration in seconds the game server needs to terminate gracefully.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"maintenance_seconds": schema.Int64Attribute{
						Description:         "The duration in seconds the game server needs to terminate gracefully when put into maintenance.",
						MarkdownDescription: "The duration in seconds the game server needs to terminate gracefully when put into maintenance.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"spec_change_seconds": schema.Int64Attribute{
						Description:         "The duration in seconds the game server needs to terminate gracefully after Formation specs have changed.",
						MarkdownDescription: "The duration in seconds the game server needs to terminate gracefully after Formation specs have changed.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"user_initiated_seconds": schema.Int64Attribute{
						Description:         "The duration in seconds the game server needs to terminate gracefully after a user has initiated a termination.",
						MarkdownDescription: "The duration in seconds the game server needs to terminate gracefully after a user has initiated a termination.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
				},
			},
			"volumes": schema.ListNestedAttribute{
				Description:         "Volumes is a list of volumes that can be mounted by containers belonging to the game server.",
				MarkdownDescription: "Volumes is a list of volumes that can be mounted by containers belonging to the game server.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the volume.",
							MarkdownDescription: "Name is the name of the volume.",
							Required:            true,
							Validators: []validator.String{
								validators.NameValidator{},
							},
						},
						"empty_dir": schema.SingleNestedAttribute{
							Description:         "EmptyDir represents a volume that is initially empty.",
							MarkdownDescription: "EmptyDir represents a volume that is initially empty.",
							Optional:            true,
							Validators: []validator.Object{
								objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("persistent")),
							},
							Attributes: map[string]schema.Attribute{
								"size_limit": schema.StringAttribute{
									Description:         "SizeLimit is the total amount of local storage required for this EmptyDir volume.",
									MarkdownDescription: "SizeLimit is the total amount of local storage required for this EmptyDir volume.",
									Optional:            true,
									Validators: []validator.String{
										validators.QuantityValidator{},
									},
								},
							},
						},
						"persistent": schema.SingleNestedAttribute{
							Description:         "Persistent represents a volume with data persistence.",
							MarkdownDescription: "Persistent represents a volume with data persistence.",
							Optional:            true,
							Validators: []validator.Object{
								objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("empty_dir")),
							},
							Attributes: map[string]schema.Attribute{
								"volume_name": schema.StringAttribute{
									Description:         "VolumeName is the name of the persistent volume.",
									MarkdownDescription: "VolumeName is the name of the persistent volume.",
									Optional:            true,
									Validators: []validator.String{
										validators.NameValidator{},
									},
								},
							},
						},
					},
				},
			},
			"gateway_policies": schema.ListAttribute{
				Description:         "GatewayPolicies is a list of gateway policies to apply to the Formation.",
				MarkdownDescription: "GatewayPolicies is a list of gateway policies to apply to the Formation.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validators.NameValidator{},
						validators.GFFieldString(formationValidator, "spec.template.spec.gatewayPolicies[?]"),
					),
				},
			},
			"profiling_enabled": schema.BoolAttribute{
				Description:         "ProfilingEnabled indicates whether profiling is enabled for the Formation.",
				MarkdownDescription: "ProfilingEnabled indicates whether profiling is enabled for the Formation.",
				Optional:            true,
			},
			"image_updater_target": schema.SingleNestedAttribute{
				Description:         "ImageUpdaterTarget is the reference that an image updater can target to match the Formation.",
				MarkdownDescription: "ImageUpdaterTarget is the reference that an image updater can target to match the Formation.",
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
func (r *formation) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *formation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan formationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.FormationV1().Formations(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Formation",
			fmt.Sprintf("Could not create Formation: %v", err),
		)
		return
	}

	plan = newFormationModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *formation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state formationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.FormationV1().Formations(state.Environment.ValueString()).Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Formation",
				fmt.Sprintf("Could not read Formation %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newFormationModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *formation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state formationModel
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
			"Error Creating Formation Patch",
			fmt.Sprintf("Could not create patch for Formation: %v", err),
		)
		return
	}

	if _, err = r.clientSet.FormationV1().Formations(newObj.Environment).Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Formation",
			fmt.Sprintf("Could not patch for Formation: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	plan.ImageUpdaterTarget = container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeFormation, oldObj.Name, oldObj.Environment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *formation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state formationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.FormationV1().Formations(state.Environment.ValueString()).Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Formation",
			fmt.Sprintf("Could not delete for Formation: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.FormationV1().Formations(state.Environment.ValueString()), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Formation Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Formation: %v", err),
		)
		return
	}
}

func (r *formation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(req.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), env)...)
}
