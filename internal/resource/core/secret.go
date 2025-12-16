package core

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	secretreg "github.com/gamefabric/gf-core/pkg/apiserver/registry/core/secret"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &secret{}
	_ resource.ResourceWithConfigure = &secret{}
)

var secretValidator = validators.NewGameFabricValidator[*corev1.Secret, secretModel](func() validators.StoreValidator {
	storage, _ := secretreg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy()
})

type secret struct {
	clientSet clientset.Interface
}

// NewSecret returns a new instance of the secret resource.
func NewSecret() resource.Resource {
	return &secret{}
}

// Metadata defines the resource type name.
func (r *secret) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

// Schema defines the schema for this resource.
func (r *secret) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
					validators.GFFieldString(secretValidator, "metadata.name"),
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
				Description:         "Description is the optional description of the secret.",
				MarkdownDescription: "Description is the optional description of the secret.",
				Optional:            true,
			},
			"data": schema.MapAttribute{
				Description:         "Data contains the secret key-value pairs. These are stored in state and persisted. Use data_wo for write-only ephemeral values.",
				MarkdownDescription: "Data contains the secret key-value pairs. These are stored in state and persisted. Use data_wo for write-only ephemeral values.",
				Optional:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.GFFieldMap(secretValidator, "data"),
					mapvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("data_wo")),
				},
			},
			"data_wo": schema.MapAttribute{
				Description:         "DataWO is a write-only version of data. Values are sensitive and write-only - they are only transmitted to the server and never displayed or stored in state.",
				MarkdownDescription: "DataWO is a write-only version of data. Values are sensitive and write-only - they are only transmitted to the server and never displayed or stored in state.",
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					mapvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("data")),
					mapvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("data_wo_version")),
				},
			},
			"data_wo_version": schema.Int64Attribute{
				Description:         "DataWOVersion is the version of the write-only data. This is used to force updates when using data_wo.",
				MarkdownDescription: "DataWOVersion is the version of the write-only data. This is used to force updates when using data_wo.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("data_wo")),
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *secret) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secret) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We must use the config, because write-only data is not part of the plan.
	var config secretModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.DataWO = config.DataWO

	obj := plan.ToObject()
	outObj, err := r.clientSet.CoreV1().Secrets(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Secret",
			fmt.Sprintf("Could not create Secret: %v", err),
		)
		return
	}

	plan = newSecretModel(outObj, config.DataWOVersion.ValueInt64())

	// Preserve plan's data as the API returns masked secrets.
	if !config.DataWOVersion.IsNull() && config.DataWOVersion.ValueInt64() > 0 {
		// Using data_wo - don't persist data in state (write-only)
		plan.Data = nil
	} else {
		// Using data - preserve the actual config values
		plan.Data = config.Data
	}

	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secret) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store current data values before reading from API (API returns masked)
	currentData := state.Data
	currentDataWOVersion := state.DataWOVersion

	outObj, err := r.clientSet.CoreV1().Secrets(state.Environment.ValueString()).Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Secret",
				fmt.Sprintf("Could not read Secret %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newSecretModel(outObj, currentDataWOVersion.ValueInt64())

	// Handle data based on whether using data_wo or regular data
	switch {
	case currentDataWOVersion.ValueInt64() > 0:
		// Using data_wo - don't persist any data in state (write-only)
		state.Data = nil
	case currentData != nil:
		// Using regular data - merge API keys with state values to detect drift
		newData := make(map[string]types.String, len(outObj.Data))
		for apiKey := range outObj.Data {
			// If key exists in both API and state, preserve the state value
			stateVal, exists := currentData[apiKey]
			if !exists {
				stateVal = types.StringNull()
			}
			newData[apiKey] = stateVal
		}
		state.Data = newData
	}

	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secret) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state secretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldObj := state.ToObject()
	newObj := plan.ToObject()

	secret, err := r.clientSet.CoreV1().Secrets(newObj.Environment).Get(ctx, newObj.Name, metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Secret for Update",
			fmt.Sprintf("Could not read Secret %q for update: %v", newObj.Name, err),
		)
		return
	}

	keys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		keys = append(keys, k)
	}

	// Remove any keys that are no longer present
	for _, k := range keys {
		if _, ok := chooseData(config)[k]; !ok {
			newObj.Data[k] = ""
		}
	}

	newObj.Data = conv.ForEachMapItem(chooseData(config), func(item types.String) string { return item.ValueString() })

	pb, err := patch.Create(oldObj, newObj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Secret Patch",
			fmt.Sprintf("Could not create patch for Secret: %v", err),
		)
		return
	}

	outObj, err := r.clientSet.CoreV1().Secrets(newObj.Environment).Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Secret",
			fmt.Sprintf("Could not patch Secret: %v", err),
		)
		return
	}

	updated := newSecretModel(outObj, plan.DataWOVersion.ValueInt64())

	updated.Data = nil
	if config.DataWOVersion.ValueInt64() == 0 {
		updated.Data = config.Data
	}

	resp.Diagnostics.Append(normalize.Model(ctx, &updated, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *secret) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.CoreV1().Secrets(state.Environment.ValueString()).Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Secret",
			fmt.Sprintf("Could not delete Secret: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.CoreV1().Secrets(state.Environment.ValueString()), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Secret Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Secret: %v", err),
		)
		return
	}
}
