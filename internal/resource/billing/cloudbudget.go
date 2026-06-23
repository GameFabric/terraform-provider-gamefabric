package billing

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	billingv2alpha1 "github.com/gamefabric/gf-core/pkg/api/billing/v2alpha1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	cloudbudgetreg "github.com/gamefabric/gf-core/pkg/apiserver/registry/billing/cloudbudget"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &cloudBudget{}
	_ resource.ResourceWithConfigure   = &cloudBudget{}
	_ resource.ResourceWithImportState = &cloudBudget{}
)

var cloudBudgetValidator = validators.NewGameFabricValidator[*billingv2alpha1.CloudBudget, cloudBudgetModel](func() validators.StoreValidator {
	storage, _ := cloudbudgetreg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

type cloudBudget struct {
	clientSet clientset.Interface
}

// NewCloudBudgetResource returns a new instance of the cloud budget resource.
func NewCloudBudgetResource() resource.Resource {
	return &cloudBudget{}
}

// Metadata defines the resource type name.
func (r *cloudBudget) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudbudget"
}

// Schema defines the schema for this resource.
func (r *cloudBudget) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Resource for managing Cloud Budgets in GameFabric Billing.",
		MarkdownDescription: "Resource for managing Cloud Budgets in GameFabric Billing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique Terraform identifier.",
				MarkdownDescription: "The unique Terraform identifier.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
				MarkdownDescription: "The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
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
					&validators.LabelsValidator{},
				},
			},
			"annotations": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					&validators.AnnotationsValidator{},
				},
			},
			"suspended": schema.BoolAttribute{
				Description:         "Suspends the cloud budget and suppresses all further notifications.",
				MarkdownDescription: "Suspends the cloud budget and suppresses all further notifications.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"receivers": schema.ListAttribute{
				Description:         "The list of notification receiver names to notify when a threshold is crossed.",
				MarkdownDescription: "The list of notification receiver names to notify when a threshold is crossed.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					validators.GFFieldList(cloudBudgetValidator, "spec.receivers"),
				},
			},
			"max_budget": schema.Float64Attribute{
				Description:         "The maximum cloud spend budget in USD. Exceeding the budget triggers an alert, it does not prevent further costs.",
				MarkdownDescription: "The maximum cloud spend budget in USD. Exceeding the budget triggers an alert, it does not prevent further costs.",
				Required:            true,
				Validators: []validator.Float64{
					validators.GFFieldFloat64(cloudBudgetValidator, "spec.maxBudget"),
				},
			},
			"thresholds": schema.ListAttribute{
				Description:         "An optional list of spend thresholds that trigger notifications, in ascending order. Values may be expressed as percentages (e.g. \"50%\") or absolute USD amounts (e.g. \"50000\"). All entries must use the same format.",
				MarkdownDescription: "An optional list of spend thresholds that trigger notifications, in ascending order. Values may be expressed as percentages (e.g. `\"50%\"`) or absolute USD amounts (e.g. `\"50000\"`). All entries must use the same format.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					validators.GFFieldList(cloudBudgetValidator, "spec.thresholds"),
				},
			},
			"interval": schema.SingleNestedAttribute{
				Description:         "Defines a start point and repetition step that are expanded into notification thresholds up to max_budget. Both start and step must use the same format: either both as percentages (e.g. \"50%\") or both as absolute USD amounts (e.g. \"50000\").",
				MarkdownDescription: "Defines a start point and repetition step that are expanded into notification thresholds up to `max_budget`. Both `start` and `step` must use the same format: either both as percentages (e.g. `\"50%\"`) or both as absolute USD amounts (e.g. `\"50000\"`).",
				Optional:            true,
				Validators: []validator.Object{
					validators.GFFieldObject(cloudBudgetValidator, "spec.interval"),
				},
				Attributes: map[string]schema.Attribute{
					"start": schema.StringAttribute{
						Description:         "The first threshold at which to begin notifying. May be a percentage (e.g. \"50%\") or an absolute USD amount (e.g. \"50000\").",
						MarkdownDescription: "The first threshold at which to begin notifying. May be a percentage (e.g. `\"50%\"`) or an absolute USD amount (e.g. `\"50000\"`).",
						Required:            true,
						Validators: []validator.String{
							validators.GFFieldString(cloudBudgetValidator, "spec.interval.start"),
						},
					},
					"step": schema.StringAttribute{
						Description:         "The step size between subsequent thresholds. May be a percentage (e.g. \"10%\") or an absolute USD amount (e.g. \"10000\").",
						MarkdownDescription: "The step size between subsequent thresholds. May be a percentage (e.g. `\"10%\"`) or an absolute USD amount (e.g. `\"10000\"`).",
						Required:            true,
						Validators: []validator.String{
							validators.GFFieldString(cloudBudgetValidator, "spec.interval.step"),
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *cloudBudget) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cloudBudget) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudBudgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.BillingV2Alpha1().CloudBudgets().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Cloud Budget",
			fmt.Sprintf("Could not create Cloud Budget: %v", err),
		)
		return
	}

	plan = newCloudBudgetModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudBudget) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudBudgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.BillingV2Alpha1().CloudBudgets().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Cloud Budget",
				fmt.Sprintf("Could not read Cloud Budget %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newCloudBudgetModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudBudget) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cloudBudgetModel
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
			"Error Creating Cloud Budget Patch",
			fmt.Sprintf("Could not create patch for Cloud Budget: %v", err),
		)
		return
	}

	if _, err = r.clientSet.BillingV2Alpha1().CloudBudgets().Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Cloud Budget",
			fmt.Sprintf("Could not patch Cloud Budget: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(newObj.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudBudget) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudBudgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.BillingV2Alpha1().CloudBudgets().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Cloud Budget",
			fmt.Sprintf("Could not delete Cloud Budget: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.BillingV2Alpha1().CloudBudgets(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Cloud Budget Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Cloud Budget: %v", err),
		)
		return
	}
}

func (r *cloudBudget) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
