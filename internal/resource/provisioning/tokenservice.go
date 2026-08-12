package provisioning

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connectTokenTypeDefault is the default value for eos.token_types: ["connect"],
// matching the GameFabric Web UI's behavior.
var connectTokenTypeDefault = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("connect")})

var (
	_ resource.Resource                   = &tokenService{}
	_ resource.ResourceWithConfigure      = &tokenService{}
	_ resource.ResourceWithImportState    = &tokenService{}
	_ resource.ResourceWithValidateConfig = &tokenService{}
)

// tokenService is the token service resource.
type tokenService struct {
	clientSet clientset.Interface
}

// NewTokenService returns a new instance of the token service resource.
func NewTokenService() resource.Resource {
	return &tokenService{}
}

// Metadata defines the resource type name.
func (r *tokenService) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_steelshield_tokenservice"
}

// Schema defines the schema for this resource.
func (r *tokenService) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// TODO - https://marbis-cloud.atlassian.net/browse/STS-2848 - share validations from GCAP
	resp.Schema = schema.Schema{
		Description: "Resource for managing Token Services. Token Service receives game client " +
			"JWTs and sends back SteelShield Tokens to game clients in exchange.",
		MarkdownDescription: "Resource for managing Token Services. Token Service receives game client " +
			"JWTs and sends back SteelShield Tokens to game clients in exchange.",
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
			"development_mode": schema.BoolAttribute{
				Description: "Whether the token service runs in development mode. When true, the token service is " +
					"provisioned in development mode and returns errors to game clients; when false, it runs in " +
					"production mode and hides errors. Defaults to false.",
				MarkdownDescription: "Whether the token service runs in development mode. When `true`, the token " +
					"service is provisioned in development mode and returns errors to game clients; when `false`, it " +
					"runs in production mode and hides errors. Defaults to `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
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
			"game_name": schema.StringAttribute{
				Description:         "The name of the game. Must contain only lowercase letters, numbers, and hyphens. Maximum length is 20 characters. Immutable after creation.",
				MarkdownDescription: "The name of the game. Must contain only lowercase letters, numbers, and hyphens. Maximum length is 20 characters. Immutable after creation.",
				Required:            true,
				Validators: []validator.String{
					validators.GameNameValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"platforms": schema.MapNestedAttribute{
				Description: "Maps platform names to game client token verification keys. Required (non-empty) " +
					"unless eos is set. When jwks is set, platforms must still be present but its " +
					"entries must not set game_client_token_keys. Must be one of: " +
					"android, ios, pc, playstation, ps4, switch, xbox.",
				MarkdownDescription: "Maps platform names to game client token verification keys. Required " +
					"(non-empty) unless `eos` is set. When `jwks` is set, platforms must still be present but " +
					"its entries must not set `game_client_token_keys`. Must be one of: " +
					"`android`, `ios`, `pc`, `playstation`, `ps4`, `switch`, `xbox`.",
				Optional: true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.OneOf(validators.TokenServiceAllowedPlatformNames...)),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"game_client_token_keys": schema.ListNestedAttribute{
							Description:         "A list of keys used to verify game client tokens for this platform.",
							MarkdownDescription: "A list of keys used to verify game client tokens for this platform.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Description: "The key material used to verify tokens. Sensitive by default, " +
											"since HMAC shared secrets are genuinely secret (only RSA/EC public keys " +
											"would not be).",
										MarkdownDescription: "The key material used to verify tokens. Sensitive by default, " +
											"since HMAC shared secrets are genuinely secret (only RSA/EC public keys " +
											"would not be).",
										Required:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
									"signing_algorithm": schema.StringAttribute{
										Description:         "The signing algorithm used with this key. Must be one of: HS256, HS384, HS512, RS256, RS384, RS512, PS256, PS384, PS512, ES256, ES384, ES512.",
										MarkdownDescription: "The signing algorithm used with this key. Must be one of: `HS256`, `HS384`, `HS512`, `RS256`, `RS384`, `RS512`, `PS256`, `PS384`, `PS512`, `ES256`, `ES384`, `ES512`.",
										Required:            true,
										Validators: []validator.String{
											stringvalidator.OneOf(validators.TokenServiceSigningAlgorithms...),
										},
									},
								},
								Validators: []validator.Object{
									validators.TokenServiceKeyMaterialValidator{},
								},
							},
						},
					},
				},
			},
			"eos": schema.ListNestedAttribute{
				Description: "Epic Online Services (EOS) auth mode configurations. Mutually exclusive with " +
					"platforms with game_client_token_keys, and jwks.",
				MarkdownDescription: "Epic Online Services (EOS) auth mode configurations. Mutually exclusive with " +
					"`platforms` with `game_client_token_keys`, and `jwks`.",
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_id": schema.StringAttribute{
							Description:         "The EOS client ID.",
							MarkdownDescription: "The EOS client ID.",
							Required:            true,
						},
						"token_types": schema.ListAttribute{
							Description: "The accepted EOS token types. Defaults to [\"connect\"], matching the Web UI " +
								"behavior. \"auth\" (EOS Auth Tokens) is accepted by the API but not supported or " +
								"recommended; use \"connect\" instead.",
							MarkdownDescription: "The accepted EOS token types. Defaults to `[\"connect\"]`, matching the Web UI " +
								"behavior. `\"auth\"` (EOS Auth Tokens) is accepted by the API but **not supported or " +
								"recommended**; use `\"connect\"` instead.",
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default:     listdefault.StaticValue(connectTokenTypeDefault),
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.OneOf("auth", "connect")),
							},
						},
						"deployment_id": schema.StringAttribute{
							Description:         "The EOS deployment ID.",
							MarkdownDescription: "The EOS deployment ID.",
							Optional:            true,
						},
						"product_id": schema.StringAttribute{
							Description:         "The EOS product ID.",
							MarkdownDescription: "The EOS product ID.",
							Optional:            true,
						},
						"sandbox_id": schema.StringAttribute{
							Description:         "The EOS sandbox ID.",
							MarkdownDescription: "The EOS sandbox ID.",
							Optional:            true,
						},
					},
				},
			},
			"jwks": schema.SingleNestedAttribute{
				Description: "JWKS endpoint auth mode configuration. Mutually exclusive with " +
					"platforms with game_client_token_keys, and eos.",
				MarkdownDescription: "JWKS endpoint auth mode configuration. Mutually exclusive with " +
					"`platforms` with `game_client_token_keys`, and `eos`.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description:         "The JWKS endpoint URL.",
						MarkdownDescription: "The JWKS endpoint URL.",
						Required:            true,
					},
				},
			},
			"state": schema.StringAttribute{
				Description:         "The high-level lifecycle state of the token service. One of \"Pending\", \"Available\", or \"Error\".",
				MarkdownDescription: "The high-level lifecycle state of the token service. One of `Pending`, `Available`, or `Error`.",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				Description:         "The fully-qualified domain name assigned to this token service instance.",
				MarkdownDescription: "The fully-qualified domain name assigned to this token service instance.",
				Computed:            true,
			},
			"platform_key": schema.StringAttribute{
				Description:         "The platform key used for all platforms.",
				MarkdownDescription: "The platform key used for all platforms.",
				Computed:            true,
				Sensitive:           true,
			},
			"reason": schema.StringAttribute{
				Description:         "The failure reason. Null when there is no failure.",
				MarkdownDescription: "The failure reason. Null when there is no failure.",
				Computed:            true,
			},
			"state_last_changed": schema.StringAttribute{
				Description:         "The timestamp of the last state change, in RFC3339 format.",
				MarkdownDescription: "The timestamp of the last state change, in RFC3339 format.",
				Computed:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *tokenService) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tokenService) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tokenServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.ProvisioningV1Beta1().TokenServices().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Token Service",
			fmt.Sprintf("Could not create Token Service: %v", err),
		)
		return
	}

	plan = newTokenServiceModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tokenService) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tokenServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outObj, err := r.clientSet.ProvisioningV1Beta1().TokenServices().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Token Service",
				fmt.Sprintf("Could not read Token Service %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newTokenServiceModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tokenService) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tokenServiceModel
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
			"Error Creating Token Service Patch",
			fmt.Sprintf("Could not create patch for Token Service: %v", err),
		)
		return
	}

	outObj, err := r.clientSet.ProvisioningV1Beta1().TokenServices().Patch(ctx, newObj.Name, rest.MergePatchType, pb, metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Patching Token Service",
			fmt.Sprintf("Could not patch Token Service: %v", err),
		)
		return
	}

	plan = newTokenServiceModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tokenService) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tokenServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.clientSet.ProvisioningV1Beta1().TokenServices().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Token Service",
			fmt.Sprintf("Could not delete Token Service: %v", err),
		)
		return
	}

	if err = wait.PollUntilNotFound(ctx, r.clientSet.ProvisioningV1Beta1().TokenServices(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Token Service Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Token Service: %v", err),
		)
		return
	}
}

func (r *tokenService) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ValidateConfig enforces that exactly one auth mode is configured: custom signing keys
// (at least one platform with game_client_token_keys), jwks, or eos. This mirrors GCAP's
// server-side strategy.validatePlatformSpecs, giving plan-time feedback instead of waiting
// for an API round trip. The backend remains the final authority.
func (r *tokenService) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if !req.Config.Raw.IsFullyKnown() {
		// Unknown values (e.g. depending on another resource): defer validation to apply time.
		return
	}

	var config tokenServiceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateTokenServiceAuthMode(config, &resp.Diagnostics)
}

// validateTokenServiceAuthMode enforces GCAP's three auth mode combinations:
// 1. Custom platform keys: platforms has keys, no jwks, no eos
// 2. JWKS: platforms present but no keys, jwks set, no eos
// 3. EOS: no jwks, eos set, platforms empty or just "eos".
func validateTokenServiceAuthMode(config tokenServiceModel, diags *diag.Diagnostics) {
	hasEOS := len(config.EOS) > 0
	hasJWKS := config.JWKS != nil
	hasPlatformKeys := tokenServiceHasPlatformKeys(config)
	hasPlatforms := len(config.Platforms) > 0

	if conflict := tokenServiceAuthModeConflict(hasEOS, hasJWKS, hasPlatformKeys); conflict != "" {
		diags.AddError("Conflicting Auth Mode Configuration", conflict)
		return
	}

	switch {
	case hasJWKS && !hasPlatforms:
		diags.AddAttributeError(
			path.Root("platforms"),
			"Missing Platforms Configuration",
			"platforms must be present when jwks is configured (without game_client_token_keys).",
		)
	case !hasEOS && !hasJWKS && !hasPlatformKeys:
		diags.AddError(
			"Missing Auth Mode Configuration",
			"Exactly one of the following must be configured: platforms with game_client_token_keys, "+
				"jwks (with platforms and no keys), or eos.",
		)
	case !hasEOS && !hasJWKS && !hasPlatforms:
		diags.AddAttributeError(
			path.Root("platforms"),
			"Missing Platforms Configuration",
			"platforms must be non-empty when neither jwks nor eos is configured.",
		)
	}
}

// tokenServiceAuthModeConflict returns a non-empty error message if more than one auth mode
// is configured at once.
func tokenServiceAuthModeConflict(hasEOS, hasJWKS, hasPlatformKeys bool) string {
	switch {
	case hasEOS && hasJWKS:
		return "eos and jwks cannot be configured together."
	case hasEOS && hasPlatformKeys:
		return "eos cannot be configured with platforms that have game_client_token_keys."
	case hasJWKS && hasPlatformKeys:
		return "jwks cannot be configured with platforms that have game_client_token_keys."
	default:
		return ""
	}
}

// tokenServiceHasPlatformKeys reports whether any platform has game_client_token_keys set.
func tokenServiceHasPlatformKeys(config tokenServiceModel) bool {
	for _, platform := range config.Platforms {
		if len(platform.GameClientTokenKeys) > 0 {
			return true
		}
	}
	return false
}
