package notification

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	notificationv1alpha1 "github.com/gamefabric/gf-core/pkg/api/notification/v1alpha1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	receiverreg "github.com/gamefabric/gf-core/pkg/apiserver/registry/notification/receiver"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &receiver{}
	_ resource.ResourceWithConfigure   = &receiver{}
	_ resource.ResourceWithImportState = &receiver{}
	_ resource.ResourceWithModifyPlan  = &receiver{}
)

var receiverValidator = validators.NewGameFabricValidator[*notificationv1alpha1.Receiver, receiverModel](func() validators.StoreValidator {
	storage, _ := receiverreg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

type receiver struct {
	clientSet clientset.Interface
}

// NewReceiverResource returns a new instance of the receiver resource.
func NewReceiverResource() resource.Resource {
	return &receiver{}
}

// Metadata defines the resource type name.
func (r *receiver) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_receiver"
}

// Schema defines the schema for this resource.
func (r *receiver) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Resource for managing Notification Receivers in GameFabric.",
		MarkdownDescription: "Resource for managing Notification Receivers in GameFabric.",
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
			"email_to": schema.ListAttribute{
				Description:         "The list of email addresses to send notifications to.",
				MarkdownDescription: "The list of email addresses to send notifications to.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					validators.GFFieldList(receiverValidator, "spec.email.to"),
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *receiver) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *receiver) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan receiverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.NotificationV1Alpha1().Receivers().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Receiver",
			fmt.Sprintf("Could not create Receiver: %v", err),
		)
		return
	}

	plan = newReceiverModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *receiver) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state receiverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.NotificationV1Alpha1().Receivers().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Receiver",
				fmt.Sprintf("Could not read Receiver %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newReceiverModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *receiver) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state receiverModel
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
			"Error Creating Receiver Patch",
			fmt.Sprintf("Could not create patch for Receiver: %v", err),
		)
		return
	}

	if _, err = r.clientSet.NotificationV1Alpha1().Receivers().Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Receiver",
			fmt.Sprintf("Could not patch Receiver: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(newObj.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *receiver) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state receiverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Before deleting the receiver, proactively remove it from any CloudBudgets that
	// still reference it. Because the receivers field is a plain string list, Terraform
	// has no dependency edge between gamefabric_cloudbudget and this resource and may
	// run the CloudBudget update and this delete in parallel. Doing the cleanup here
	// guarantees the API will accept the deletion regardless of scheduling order.
	if err := r.removeFromReferencingCloudBudgets(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Receiver Reference from CloudBudgets",
			fmt.Sprintf("Could not remove Receiver %q from referencing CloudBudgets before deletion: %v", state.Name.ValueString(), err),
		)
		return
	}

	err := r.clientSet.NotificationV1Alpha1().Receivers().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Receiver",
			fmt.Sprintf("Could not delete Receiver: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.NotificationV1Alpha1().Receivers(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Receiver Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Receiver: %v", err),
		)
		return
	}
}

// removeFromReferencingCloudBudgets removes this receiver from every CloudBudget that still lists
// it in spec.receivers. This is called during Delete so that the API accepts the deletion
// regardless of whether Terraform's concurrent CloudBudget update has already completed.
//
// If removing this receiver would leave a CloudBudget with zero receivers (which the API forbids),
// the patch is skipped for that budget. The subsequent receiver deletion will then fail at the API
// level with a clear error, prompting the user to fix the CloudBudget first.
func (r *receiver) removeFromReferencingCloudBudgets(ctx context.Context, receiverName string) error {
	budgets, err := r.clientSet.BillingV2Alpha1().CloudBudgets().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("could not list CloudBudgets: %w", err)
	}

	for i := range budgets.Items {
		budget := &budgets.Items[i]
		if !slices.Contains(budget.Spec.Receivers, receiverName) {
			continue
		}

		newReceivers := slices.DeleteFunc(slices.Clone(budget.Spec.Receivers), func(r string) bool {
			return r == receiverName
		})
		// The API requires at least one receiver. Skip the patch and let the
		// delete fail naturally so the user gets a clear API-level error.
		if len(newReceivers) == 0 {
			continue
		}

		updated := *budget
		updated.Spec.Receivers = newReceivers

		pb, err := patch.Create(budget, &updated)
		if err != nil {
			return fmt.Errorf("could not create patch for CloudBudget %q: %w", budget.Name, err)
		}

		if _, err = r.clientSet.BillingV2Alpha1().CloudBudgets().Patch(
			ctx, budget.Name, rest.MergePatchType, pb, metav1.UpdateOptions{},
		); err != nil {
			return fmt.Errorf("could not patch CloudBudget %q: %w", budget.Name, err)
		}
	}

	return nil
}

// ModifyPlan checks whether deleting this receiver would leave any CloudBudget with an empty
// receivers list, which the API forbids and which removeFromReferencingCloudBudgets skips.
// In that case we surface a plan-time warning so the user knows the apply will fail and must
// update the CloudBudget to use a different receiver first.
//
// For CloudBudgets that have additional receivers, no warning is needed: Delete will
// proactively remove this receiver from those budgets before calling the deletion API.
func (r *receiver) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only run on destroy (plan is null, state has a value).
	if !req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	// clientSet may be nil during unit tests that do not configure the provider.
	if r.clientSet == nil {
		return
	}

	var state receiverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	receiverName := state.Name.ValueString()

	budgets, err := r.clientSet.BillingV2Alpha1().CloudBudgets().List(ctx, metav1.ListOptions{})
	if err != nil {
		// Non-fatal: surface a warning so the user is aware.
		resp.Diagnostics.AddWarning(
			"Could Not Verify Receiver References",
			fmt.Sprintf("Could not list CloudBudgets to check whether Receiver %q is still in use: %v. "+
				"The apply may fail if the Receiver is the last one on a CloudBudget.", receiverName, err),
		)
		return
	}

	var budgetNames []string
	for _, budget := range budgets.Items {
		if slices.Contains(budget.Spec.Receivers, receiverName) && len(budget.Spec.Receivers) == 1 {
			budgetNames = append(budgetNames, budget.Name)
		}
	}

	if len(budgetNames) > 0 {
		resp.Diagnostics.AddWarning(
			"Receiver Is the Last Reference on a CloudBudget",
			fmt.Sprintf(
				"Deleting Receiver %q will fail because it is the only receiver on CloudBudget(s): %s.\n\n"+
					"Update those CloudBudget(s) to reference a different receiver before deleting this one.",
				receiverName, strings.Join(budgetNames, ", "),
			),
		)
	}
}

func (r *receiver) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
