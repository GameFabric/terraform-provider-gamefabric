package authentication

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-apiclient/tools/patch"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apiserver/registry/generic"
	authenticationv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	providerreg "github.com/gamefabric/gf-core/pkg/apiserver/registry/authentication/provider"
	"github.com/gamefabric/gf-core/pkg/apiserver/registry/registrytest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &provider{}
	_ resource.ResourceWithConfigure   = &provider{}
	_ resource.ResourceWithImportState = &provider{}
)

var providerValidator = validators.NewGameFabricValidator[*authenticationv1.Provider, providerModel](func() validators.StoreValidator {
	storage, _ := providerreg.New(generic.StoreOptions{Config: generic.Config{
		StorageFactory: registrytest.FakeStorageFactory{},
	}})
	return storage.Store.Strategy
})

// provider is the provider resource.
type provider struct {
	clientSet clientset.Interface
}

// NewProvider creates a new provider resource.
func NewProvider() resource.Resource {
	return &provider{}
}

// Metadata returns the resource metadata.
func (r *provider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authentication_provider"
}

// Schema defines the schema for the provider resource.
func (r *provider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					validators.GFFieldString(providerValidator, "metadata.name"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description:         "The user-friendly name of the provider.",
				MarkdownDescription: "The user-friendly name of the provider.",
				Optional:            true,
			},
			"oidc": schema.SingleNestedAttribute{
				Description:         "OIDC (OpenID Connect) provider configuration for the authentication provider.",
				MarkdownDescription: "OIDC (OpenID Connect) provider configuration for the authentication provider.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Description:         "ClientID is the client ID of the OIDC provider.",
						MarkdownDescription: "ClientID is the client ID of the OIDC provider.",
						Required:            true,
					},
					"client_secret": schema.StringAttribute{
						Description:         "ClientSecret is the client secret of the OIDC provider.",
						MarkdownDescription: "ClientSecret is the client secret of the OIDC provider.",
						Required:            true,
						Sensitive:           true,
					},
					"issuer": schema.StringAttribute{
						Description:         "Issuer is the issuer URL of the OIDC provider.",
						MarkdownDescription: "Issuer is the issuer URL of the OIDC provider.",
						Required:            true,
					},
					"redirect_uri": schema.StringAttribute{
						Description:         "RedirectURI is the redirect URI of the OIDC provider.",
						MarkdownDescription: "RedirectURI is the redirect URI of the OIDC provider.",
						Required:            true,
					},
					"acr_values": schema.ListAttribute{
						Description: "ACRValues (Authentication Context Class Reference Values) that specifies the Authentication Context Class " +
							"Values within the Authentication Request that the Authorization Server is being requested to use for " +
							"processing requests from this Client, with the values appearing in order of preference.",
						MarkdownDescription: "ACRValues (Authentication Context Class Reference Values) that specifies the Authentication Context Class " +
							"Values within the Authentication Request that the Authorization Server is being requested to use for " +
							"processing requests from this Client, with the values appearing in order of preference.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"allowed_groups": schema.ListAttribute{
						Description:         "AllowedGroups is a list of groups that are allowed to authenticate with this provider.",
						MarkdownDescription: "AllowedGroups is a list of groups that are allowed to authenticate with this provider.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"basic_auth_unsupported": schema.BoolAttribute{
						Description: "Basic auth unsupported causes client_secret to be passed as POST parameters instead of basic auth. " +
							"This is specifically \"NOT RECOMMENDED\" by the OAuth2 RFC, but some providers require it.\n\n" +
							"https://tools.ietf.org/html/rfc6749#section-2.3.1",
						MarkdownDescription: "Basic auth unsupported causes client_secret to be passed as POST parameters instead of basic auth. " +
							"This is specifically \"NOT RECOMMENDED\" by the OAuth2 RFC, but some providers require it.\n\n" +
							"https://tools.ietf.org/html/rfc6749#section-2.3.1",
						Optional: true,
					},
					"claim_mapping": schema.SingleNestedAttribute{
						Description:         "Claim mapping contains the claim mapping options.",
						MarkdownDescription: "Claim mapping contains the claim mapping options.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"preferred_username": schema.StringAttribute{
								Description:         "Configurable key which contains the preferred username claims. Defaults to \"preferred_username\".",
								MarkdownDescription: "Configurable key which contains the preferred username claims. Defaults to \"preferred_username\".",
								Optional:            true,
							},
							"email": schema.StringAttribute{
								Description:         "Configurable key which contains the email claims. Defaults to \"email\".",
								MarkdownDescription: "Configurable key which contains the email claims. Defaults to \"email\".",
								Optional:            true,
							},
							"groups": schema.StringAttribute{
								Description:         "Configurable key which contains the groups claims. Defaults to \"groups\".",
								MarkdownDescription: "Configurable key which contains the groups claims. Defaults to \"groups\".",
								Optional:            true,
							},
						},
					},
					"claim_modification": schema.SingleNestedAttribute{
						Description:         "Claim modification contains all claim mutations options.",
						MarkdownDescription: "Claim modification contains all claim mutations options.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"filter_group_claims": schema.SingleNestedAttribute{
								Description: "Filter group claims is a regex filter used to keep only the matching groups. " +
									"This is useful when the groups list is too large to fit within an HTTP header.",
								MarkdownDescription: "Filter group claims is a regex filter used to keep only the matching groups. " +
									"This is useful when the groups list is too large to fit within an HTTP header.",
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"groups_filter": schema.StringAttribute{
										Description:         "Groups filter is the regex filter used to keep only the matching groups.",
										MarkdownDescription: "Groups filter is the regex filter used to keep only the matching groups.",
										Optional:            true,
									},
								},
							},
							"new_group_from_claims": schema.ListNestedAttribute{
								Description:         "New group from claims specifies how claims can be joined to create groups.",
								MarkdownDescription: "New group from claims specifies how claims can be joined to create groups.",
								Optional:            true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"claims": schema.ListAttribute{
											Description:         "Claims is a list of claim to join together.",
											MarkdownDescription: "Claims is a list of claim to join together.",
											Optional:            true,
											ElementType:         types.StringType,
										},
										"clear_delimiter": schema.BoolAttribute{
											Description:         "Clear delimiter indicates if the Delimiter string should be removed from claim values.",
											MarkdownDescription: "Clear delimiter indicates if the Delimiter string should be removed from claim values.",
											Optional:            true,
										},
										"delimiter": schema.StringAttribute{
											Description:         "Delimiter is the string used to separate the claims.",
											MarkdownDescription: "Delimiter is the string used to separate the claims.",
											Optional:            true,
										},
										"prefix": schema.StringAttribute{
											Description:         "Prefix is a string to place before the first claim.",
											MarkdownDescription: "Prefix is a string to place before the first claim.",
											Optional:            true,
										},
									},
								},
							},
						},
					},
					"get_user_info": schema.BoolAttribute{
						Description: "GetUserInfo uses the userinfo endpoint to get additional claims for the token. " +
							"This is especially useful where upstreams return \"thin\" id tokens.",
						MarkdownDescription: "GetUserInfo uses the userinfo endpoint to get additional claims for the token. " +
							"This is especially useful where upstreams return \"thin\" id tokens.",
						Optional: true,
					},
					"hosted_domains": schema.ListAttribute{
						Description: "HostedDomains was an optional list of whitelisted domains when using the OIDC provider with Google. " +
							"Only users from a whitelisted domain were allowed to log in. " +
							"Support for this option was removed from the OIDC provider. " +
							"Consider switching to the Google provider which supports this option.",
						MarkdownDescription: "HostedDomains was an optional list of whitelisted domains when using the OIDC provider with Google. " +
							"Only users from a whitelisted domain were allowed to log in. " +
							"Support for this option was removed from the OIDC provider. " +
							"Consider switching to the Google provider which supports this option.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"insecure_enable_groups": schema.BoolAttribute{
						Description:         "InsecureEnableGroups enables groups claims.",
						MarkdownDescription: "InsecureEnableGroups enables groups claims.",
						Optional:            true,
					},
					"insecure_skip_email_verified": schema.BoolAttribute{
						Description:         "Insecure skip email verified overrides the value of email_verified to true in the returned claims.",
						MarkdownDescription: "Insecure skip email verified overrides the value of email_verified to true in the returned claims.",
						Optional:            true,
					},
					"insecure_skip_verify": schema.BoolAttribute{
						Description:         "Insecure skip verify disabled certificate verification. Use with caution.",
						MarkdownDescription: "Insecure skip verify disabled certificate verification. Use with caution.",
						Optional:            true,
					},
					"override_claim_mapping": schema.BoolAttribute{
						Description: "OverrideClaimMapping will be used to override the options defined in claimMappings. " +
							"i.e. if there are 'email' and `preferred_email` claims available, " +
							"by default Dex will always use the `email` claim independent of the ClaimMapping.EmailKey. " +
							"This setting allows you to override the default behavior of Dex and enforce the mappings defined in `claimMapping`.",
						MarkdownDescription: "OverrideClaimMapping will be used to override the options defined in claimMappings. " +
							"i.e. if there are 'email' and `preferred_email` claims available, " +
							"by default Dex will always use the `email` claim independent of the ClaimMapping.EmailKey. " +
							"This setting allows you to override the default behavior of Dex and enforce the mappings defined in `claimMapping`.",
						Optional: true,
					},
					"prompt_type": schema.StringAttribute{
						Description:         "Prompt type will be used for the prompt parameter. When offline_access scope is used this defaults to prompt=consent.",
						MarkdownDescription: "Prompt type will be used for the prompt parameter. When offline_access scope is used this defaults to prompt=consent.",
						Optional:            true,
					},
					"provider_discovery_override": schema.SingleNestedAttribute{
						Description:         "OIDC provider discovery overrides contains the options to override the discovered urls.",
						MarkdownDescription: "OIDC provider discovery overrides contains the options to override the discovered urls.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"auth_url": schema.StringAttribute{
								Description: "AuthURL provides a way to user overwrite the Auth URL " +
									"from the .well-known/openid-configuration authorization_endpoint.",
								MarkdownDescription: "AuthURL provides a way to user overwrite the Auth URL " +
									"from the .well-known/openid-configuration authorization_endpoint.",
								Optional: true,
							},
							"jwks_url": schema.StringAttribute{
								Description: "JWKSURL provides a way to user overwrite the JWKS URL " +
									"from the .well-known/openid-configuration jwks_uri.",
								MarkdownDescription: "JWKSURL provides a way to user overwrite the JWKS URL " +
									"from the .well-known/openid-configuration jwks_uri.",
								Optional: true,
							},
							"token_url": schema.StringAttribute{
								Description: "TokenURL provides a way to user overwrite the Token URL " +
									"from the .well-known/openid-configuration token_endpoint.",
								MarkdownDescription: "TokenURL provides a way to user overwrite the Token URL " +
									"from the .well-known/openid-configuration token_endpoint.",
								Optional: true,
							},
						},
					},
					"root_ca_certificates": schema.ListAttribute{
						Description:         "Root CAs are root certificates for SSL validation.",
						MarkdownDescription: "Root CAs are root certificates for SSL validation.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"scopes": schema.ListAttribute{
						Description:         "Scopes are the scopes requested from the OIDC provider. Defaults to \"profile\" and \"email\".",
						MarkdownDescription: "Scopes are the scopes requested from the OIDC provider. Defaults to \"profile\" and \"email\".",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.UniqueValues(),
							listvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(
									// https://dexidp.io/docs/configuration/custom-scopes-claims-clients/
									regexp.MustCompile(`^openid|email|profile|groups|federated:id|offline_access|audience:server:client_id:\(.*\)$`),
									"must be one of these: openid, email, profile, groups, federated:id, offline_access, audience:server:client_id:(client-id).",
								),
							),
						},
					},
					"user_id_key": schema.StringAttribute{
						Description:         "User ID Key is the key used to identify the user in the claims.",
						MarkdownDescription: "User ID Key is the key used to identify the user in the claims.",
						Optional:            true,
					},
					"user_name_key": schema.StringAttribute{
						Description:         "User Name Key is the key used to identify the username in the claims.",
						MarkdownDescription: "User Name Key is the key used to identify the username in the claims.",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *provider) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	procCtx, ok := req.ProviderData.(*provcontext.Context)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *provcontext.Context, got: %T", req.ProviderData),
		)
		return
	}

	r.clientSet = procCtx.ClientSet
}

// Create creates the resource and sets the initial Terraform state on success.
func (r *provider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan providerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := plan.ToObject()
	outObj, err := r.clientSet.AuthenticationV1Beta1().Providers().Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Error",
			fmt.Sprintf("Could not create Authentication Provider %q: %v", plan.Name.ValueString(), err),
		)
		return
	}
	plan = newProviderModel(outObj)
	resp.Diagnostics.Append(normalize.Model(ctx, &plan, req.Plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *provider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state providerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.AuthenticationV1Beta1().Providers().Get(ctx, state.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Read Error",
				fmt.Sprintf("Could not read Authentication Provider %q: %v", state.Name.ValueString(), err),
			)
		}
		return
	}

	state = newProviderModel(obj)
	resp.Diagnostics.Append(normalize.Model(ctx, &state, req.State)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *provider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state providerModel
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
			"Creating Provider Patch",
			fmt.Sprintf("Could not create patch for Authentication Provider %q: %v", state.Name.ValueString(), err),
		)
		return
	}

	if _, err = r.clientSet.AuthenticationV1Beta1().Providers().Patch(ctx, state.Name.ValueString(), rest.MergePatchType, pb, metav1.UpdateOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Patching Provider",
			fmt.Sprintf("Could not patch Authentication Provider %q: %v", state.Name.ValueString(), err),
		)
		return
	}

	plan.ID = types.StringValue(newObj.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *provider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state providerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.clientSet.AuthenticationV1Beta1().Providers().Delete(ctx, state.Name.ValueString(), metav1.DeleteOptions{}); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Authentication Provider",
			fmt.Sprintf("Could not delete Authentication Provider %q: %v", state.Name.ValueString(), err),
		)
		return
	}

	if err := wait.PollUntilNotFound(ctx, r.clientSet.AuthenticationV1Beta1().Providers(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Authentication Provider Deletion",
			fmt.Sprintf("Timed out waiting for deletion of Authentication Provider: %v", err),
		)
	}
}

func (r *provider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
