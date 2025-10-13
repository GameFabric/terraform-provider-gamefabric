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
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
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
	_ resource.Resource                = &region{}
	_ resource.ResourceWithConfigure   = &region{}
	_ resource.ResourceWithImportState = &region{}
)

type region struct {
	clientSet clientset.Interface
}

func NewRegion() resource.Resource {
	return &region{}
}

func (r *region) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

// Schema defines the schema for this data source.
func (r *region) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			},
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is the user-friendly name of a region.",
				MarkdownDescription: "DisplayName is the user-friendly name of a region.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Description:         "Description is the optional description of the region.",
				MarkdownDescription: "Description is the optional description of the region.",
				Optional:            true,
			},
			"types": schema.MapNestedAttribute{
				Description:         "Types defines the types on infrastructure available in the region.",
				MarkdownDescription: "Types defines the types on infrastructure available in the region.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"locations": schema.ListAttribute{
							Description:         "Locations defines the locations for a type.",
							MarkdownDescription: "Locations defines the locations for a type.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"env": schema.ListNestedAttribute{
							Description:         "Env is a list of environment variables to set on all containers in this region.",
							MarkdownDescription: "Env is a list of environment variables to set on all containers in this region.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description:         "Name is the name of the environment variable.",
										MarkdownDescription: "Name is the name of the environment variable.",
										Required:            true,
									},
									"value": schema.StringAttribute{
										Description:         "Value is the value of the environment variable.",
										MarkdownDescription: "Value is the value of the environment variable.",
										Optional:            true,
										Validators: []validator.String{
											stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_from")),
										},
									},
									"value_from": schema.SingleNestedAttribute{
										Description:         "ValueFrom is the source for the environment variable&#39;s value.",
										MarkdownDescription: "ValueFrom is the source for the environment variable&#39;s value.",
										Optional:            true,
										Attributes: map[string]schema.Attribute{
											"field_ref": schema.SingleNestedAttribute{
												Description:         "FieldRef selects the field of the pod. Supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`, metadata.armadaName, metadata.regionName, metadata.regionTypeName, metadata.siteName, metadata.imageBranch, metadata.imageName, metadata.imageTag, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.",
												MarkdownDescription: "FieldRef selects the field of the pod. Supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`, metadata.armadaName, metadata.regionName, metadata.regionTypeName, metadata.siteName, metadata.imageBranch, metadata.imageName, metadata.imageTag, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.",
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"api_version": schema.StringAttribute{
														Required: true,
													},
													"field_path": schema.StringAttribute{
														Required: true,
													},
												},
												Validators: []validator.Object{
													objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("config_file_key_ref")),
												},
											},
											"config_file_key_ref": schema.SingleNestedAttribute{
												Description:         "ConfigFileKeyRef select the configuration file.",
												MarkdownDescription: "ConfigFileKeyRef select the configuration file.",
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description:         "Name is the name of the configuration file.",
														MarkdownDescription: "Name is the name of the configuration file.",
														Required:            true,
													},
												},
												Validators: []validator.Object{
													objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("field_ref")),
												},
											},
										},
										Validators: []validator.Object{
											objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value")),
										},
									},
								},
							},
						},
						"scheduling": schema.StringAttribute{
							Description:         "Scheduling strategy. Defaults to &#34;Packed&#34;",
							MarkdownDescription: "Scheduling strategy. Defaults to &#34;Packed&#34;",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("Packed", "Distributed"),
							},
						},
					},
				},
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *region) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	return
}

func (r *region) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan regionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	_, err := r.clientSet.CoreV1().Regions(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Region",
			fmt.Sprintf("Could not create Region: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *region) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state regionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.CoreV1().Regions(state.Environment.ValueString()).Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Region",
				fmt.Sprintf("Could not read Region %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newRegionModel(outObj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *region) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state regionModel
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
			"Error Creating Region Patch",
			fmt.Sprintf("Could not create patch for Region: %v", err),
		)
		return
	}

	if _, err = r.clientSet.CoreV1().Regions(newObj.Environment).Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Region",
			fmt.Sprintf("Could not patch for Region: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(cache.NewObjectName(newObj.Environment, newObj.Name).String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *region) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state regionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.CoreV1().Regions(state.Environment.ValueString()).Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Region",
			fmt.Sprintf("Could not delete for Region: %v", err),
		)
		return
	}

	// TODO: wait for deletion
}

func (r *region) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		return
	}

	env, name := cache.SplitMetaNamespaceKey(req.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), env)...)
}
