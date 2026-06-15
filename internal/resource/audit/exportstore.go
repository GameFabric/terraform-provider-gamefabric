package audit

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &exportStore{}
	_ resource.ResourceWithConfigure   = &exportStore{}
	_ resource.ResourceWithImportState = &exportStore{}
)

type exportStore struct {
	clientSet clientset.Interface
}

// NewExportStore returns a new instance of the exportstore resource.
func NewExportStore() resource.Resource {
	return &exportStore{}
}

// Metadata defines the resource type name.
func (r *exportStore) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exportstore"
}

// Schema defines the schema for this resource.
func (r *exportStore) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description:         "The unique object name. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
				MarkdownDescription: "The unique object name. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"suspended": schema.BoolAttribute{
				Description:         "Controls whether audit log export is stopped. Setting this to true triggers a drain: the controller flushes pending events before transitioning to Suspended.",
				MarkdownDescription: "Controls whether audit log export is stopped. Setting this to `true` triggers a drain: the controller flushes pending events before transitioning to `Suspended`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"s3": schema.SingleNestedAttribute{
				Description:         "S3-compatible storage configuration. Works with AWS S3, Cloudflare R2, and other S3-compatible stores.",
				MarkdownDescription: "S3-compatible storage configuration. Works with AWS S3, Cloudflare R2, and other S3-compatible stores.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Description:         "The bucket name.",
						MarkdownDescription: "The bucket name.",
						Required:            true,
					},
					"region": schema.StringAttribute{
						Description:         "The storage region. Defaults to \"auto\", but usually required for AWS.",
						MarkdownDescription: "The storage region. Defaults to `auto`, but usually required for AWS.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"endpoint": schema.StringAttribute{
						Description:         "The S3-compatible endpoint URL. Defaults to \"https://s3.amazonaws.com\".",
						MarkdownDescription: "The S3-compatible endpoint URL. Defaults to `https://s3.amazonaws.com`.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"prefix": schema.StringAttribute{
						Description:         "An optional key prefix (folder) within the bucket.",
						MarkdownDescription: "An optional key prefix (folder) within the bucket.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"auth": schema.SingleNestedAttribute{
						Description:         "Authentication configuration for the S3-compatible store.",
						MarkdownDescription: "Authentication configuration for the S3-compatible store.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"access_key_id": schema.StringAttribute{
								Description:         "The access key ID for static credential authentication.",
								MarkdownDescription: "The access key ID for static credential authentication.",
								Required:            true,
							},
							"secret_access_key": schema.StringAttribute{
								Description:         "The secret access key for static credential authentication. Write-only: never stored in state.",
								MarkdownDescription: "The secret access key for static credential authentication. Write-only: never stored in state.",
								Required:            true,
								Sensitive:           true,
								WriteOnly:           true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_access_key_version")),
								},
							},
							"secret_access_key_version": schema.Int64Attribute{
								Description:         "Increment this value to rotate the secret access key.",
								MarkdownDescription: "Increment this value to rotate the secret access key.",
								Required:            true,
								Validators: []validator.Int64{
									int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_access_key")),
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *exportStore) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *exportStore) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan exportStoreModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read config to access write-only secret_access_key (absent from plan).
	var config exportStoreModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	if config.S3 != nil {
		obj.Spec.S3.Auth.SecretAccessKey = config.S3.Auth.SecretAccessKey.ValueString()
	}

	outObj, err := r.clientSet.AuditV1Alpha1().ExportStores().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating ExportStore",
			fmt.Sprintf("Could not create ExportStore: %v", err),
		)
		return
	}

	plan = newExportStoreModel(outObj)
	preserveWriteOnly(&plan, config, types.Int64Null())

	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *exportStore) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state exportStoreModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve write-only version from state before overwriting.
	var savedVersion types.Int64
	if state.S3 != nil {
		savedVersion = state.S3.Auth.SecretAccessKeyVersion
	}

	outObj, err := r.clientSet.AuditV1Alpha1().ExportStores().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading ExportStore",
				fmt.Sprintf("Could not read ExportStore %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newExportStoreModel(outObj)
	if state.S3 != nil {
		state.S3.Auth.SecretAccessKeyVersion = savedVersion
	}

	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *exportStore) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state exportStoreModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read config to access write-only secret_access_key (absent from plan).
	var config exportStoreModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldObj := state.ToObject()
	// Restore masked value on old object so the patch diff is clean.
	if oldObj.Spec.S3 != nil {
		oldObj.Spec.S3.Auth.SecretAccessKey = auditv1alpha1.MaskedValue
	}

	newObj := plan.ToObject()
	switch {
	case config.S3 != nil && !config.S3.Auth.SecretAccessKey.IsNull():
		newObj.Spec.S3.Auth.SecretAccessKey = config.S3.Auth.SecretAccessKey.ValueString()
	case newObj.Spec.S3 != nil:
		// No new key provided: send masked value so server transparently restores the stored credential.
		newObj.Spec.S3.Auth.SecretAccessKey = auditv1alpha1.MaskedValue
	}

	pb, err := patch.Create(oldObj, newObj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating ExportStore Patch",
			fmt.Sprintf("Could not create patch for ExportStore: %v", err),
		)
		return
	}

	if _, err = r.clientSet.AuditV1Alpha1().ExportStores().Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Patching ExportStore",
			fmt.Sprintf("Could not patch ExportStore: %v", err),
		)
		return
	}

	plan.ID = types.StringValue(newObj.Name)
	preserveWriteOnly(&plan, config, plan.S3.Auth.SecretAccessKeyVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *exportStore) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state exportStoreModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.AuditV1Alpha1().ExportStores().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting ExportStore",
			fmt.Sprintf("Could not delete ExportStore: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.AuditV1Alpha1().ExportStores(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for ExportStore Deletion",
			fmt.Sprintf("Timed out waiting for deletion of ExportStore: %v", err),
		)
	}
}

func (r *exportStore) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// preserveWriteOnly zeroes out the write-only secret_access_key in the model
// (so it is never stored in state) and restores the version counter from the config.
func preserveWriteOnly(m *exportStoreModel, config exportStoreModel, version types.Int64) {
	if m.S3 == nil {
		return
	}
	m.S3.Auth.SecretAccessKey = types.StringNull()
	if config.S3 != nil {
		m.S3.Auth.SecretAccessKeyVersion = config.S3.Auth.SecretAccessKeyVersion
	} else {
		m.S3.Auth.SecretAccessKeyVersion = version
	}
}
