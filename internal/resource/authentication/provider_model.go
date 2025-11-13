package authentication

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type providerModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	DisplayName types.String  `tfsdk:"display_name"`
	OIDC        *oidcProvider `tfsdk:"oidc"`
}

func newProviderModel(obj *authv1.Provider) providerModel {
	return providerModel{
		ID:          types.StringValue(obj.Name),
		Name:        types.StringValue(obj.Name),
		DisplayName: types.StringValue(obj.Spec.DisplayName),
		OIDC:        newOIDCProvider(obj.Spec.OIDC),
	}
}

func (m providerModel) ToObject() *authv1.Provider {
	return &authv1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name.ValueString(),
		},
		Spec: authv1.ProviderSpec{
			DisplayName: m.DisplayName.ValueString(),
			OIDC:        m.OIDC.ToObject(),
		},
	}
}

// oidcProvider is the OIDC provider configuration.
type oidcProvider struct {
	// Issuer is the issuer URL of the OIDC provider.
	//
	//openapi:required
	Issuer types.String `tfsdk:"issuer"`

	// ClientID is the client ID of the OIDC provider.
	ClientID types.String `tfsdk:"client_id"`

	// ClientSecret is the client secret of the OIDC provider.
	ClientSecret types.String `tfsdk:"client_secret"`

	// RedirectURI is the redirect URI of the OIDC provider.
	RedirectURI types.String `tfsdk:"redirect_uri"`

	// The section to override options discovered automatically from
	// the providers' discovery URL (.well-known/openid-configuration).
	ProviderDiscoveryOverrides *oidcProviderDiscoveryOverride `tfsdk:"provider_discovery_override"`

	// BasicAuthUnsupported causes client_secret to be passed as POST parameters instead of basic
	// auth. This is specifically "NOT RECOMMENDED" by the OAuth2 RFC, but some
	// providers require it.
	//
	// https://tools.ietf.org/html/rfc6749#section-2.3.1
	BasicAuthUnsupported types.Bool `tfsdk:"basic_auth_unsupported"`

	// Scopes are the scopes requested from the OIDC provider.
	// Defaults to "profile" and "email".
	Scopes []types.String `tfsdk:"scopes"`

	// HostedDomains was an optional list of whitelisted domains when using the OIDC provider with Google.
	// Only users from a whitelisted domain were allowed to log in.
	// Support for this option was removed from the OIDC provider.
	// Consider switching to the Google provider which supports this option.
	//
	// Deprecated: will be removed in future releases.
	HostedDomains []types.String `tfsdk:"hosted_domains"`

	// RootCAs are root certificates for SSL validation.
	RootCAs []types.String `tfsdk:"root_ca_certificates"`

	// InsecureSkipEmailVerified overrides the value of email_verified to true in the returned claims.
	InsecureSkipEmailVerified types.Bool `tfsdk:"insecure_skip_email_verified"`

	// InsecureEnableGroups enables groups claims.
	InsecureEnableGroups types.Bool `tfsdk:"insecure_enable_groups"`

	// AllowedGroups is a list of groups that are allowed to authenticate with this provider.
	AllowedGroups []types.String `tfsdk:"allowed_groups"`

	// ACRValues (Authentication Context Class Reference Values) that specifies the Authentication Context Class Values
	// within the Authentication Request that the Authorization Server is being requested to use for
	// processing requests from this Client, with the values appearing in order of preference.
	ACRValues []types.String `tfsdk:"acr_values"`

	// InsecureSkipVerify disabled certificate verification. Use with caution.
	InsecureSkipVerify types.Bool `tfsdk:"insecure_skip_verify"`

	// GetUserInfo uses the userinfo endpoint to get additional claims for
	// the token. This is especially useful where upstreams return "thin"
	// id tokens.
	GetUserInfo types.Bool `tfsdk:"get_user_info"`

	// UserIDKey is the key used to identify the user in the claims.
	UserIDKey types.String `tfsdk:"user_id_key"`

	// UserNameKey is the key used to identify the username in the claims.
	UserNameKey types.String `tfsdk:"user_name_key"`

	// PromptType will be used for the prompt parameter. When offline_access scope is used this defaults to prompt=consent.
	PromptType types.String `tfsdk:"prompt_type"`

	// OverrideClaimMapping will be used to override the options defined in claimMappings.
	// i.e. if there are 'email' and `preferred_email` claims available, by default Dex will always use the `email` claim independent of the ClaimMapping.EmailKey.
	// This setting allows you to override the default behavior of Dex and enforce the mappings defined in `claimMapping`.
	OverrideClaimMapping types.Bool `tfsdk:"override_claim_mapping"`

	// ClaimMapping contains all claim mapping overrides.
	ClaimMapping *oidcProviderClaimMapping `tfsdk:"claim_mapping"`

	// ClaimMutations contains all claim mutations options.
	ClaimMutations *oidcProviderClaimMutations `tfsdk:"claim_modification"`
}

func newOIDCProvider(spec *authv1.OIDCProvider) *oidcProvider {
	return &oidcProvider{
		Issuer:                     types.StringValue(spec.Issuer),
		ClientID:                   types.StringValue(spec.ClientID),
		ClientSecret:               types.StringValue(spec.ClientSecret),
		RedirectURI:                types.StringValue(spec.RedirectURI),
		ProviderDiscoveryOverrides: newOIDCProviderDiscoveryOverride(spec.ProviderDiscoveryOverrides),
		BasicAuthUnsupported:       types.BoolPointerValue(spec.BasicAuthUnsupported),
		Scopes:                     conv.ForEachSliceItem(spec.Scopes, types.StringValue),
		HostedDomains:              conv.ForEachSliceItem(spec.HostedDomains, types.StringValue), //nolint:staticcheck // Useful for OIDC with Google.
		RootCAs:                    conv.ForEachSliceItem(spec.RootCAs, types.StringValue),
		InsecureSkipEmailVerified:  types.BoolValue(spec.InsecureSkipEmailVerified),
		InsecureEnableGroups:       types.BoolValue(spec.InsecureEnableGroups),
		AllowedGroups:              conv.ForEachSliceItem(spec.AllowedGroups, types.StringValue),
		ACRValues:                  conv.ForEachSliceItem(spec.ACRValues, types.StringValue),
		InsecureSkipVerify:         types.BoolValue(spec.InsecureSkipVerify),
		GetUserInfo:                types.BoolValue(spec.GetUserInfo),
		UserIDKey:                  types.StringValue(spec.UserIDKey),
		UserNameKey:                types.StringValue(spec.UserNameKey),
		PromptType:                 types.StringPointerValue(spec.PromptType),
		OverrideClaimMapping:       types.BoolValue(spec.OverrideClaimMapping),
		ClaimMapping:               newOIDCClaimMapping(spec.ClaimMapping),
		ClaimMutations:             newOIDCClaimMutations(spec.ClaimMutations),
	}
}

func (m *oidcProvider) ToObject() *authv1.OIDCProvider {
	if m == nil {
		return nil
	}
	return &authv1.OIDCProvider{
		Issuer:                     m.Issuer.ValueString(),
		ClientID:                   m.ClientID.ValueString(),
		ClientSecret:               m.ClientSecret.ValueString(),
		RedirectURI:                m.RedirectURI.ValueString(),
		ProviderDiscoveryOverrides: m.ProviderDiscoveryOverrides.ToObject(),
		BasicAuthUnsupported:       m.BasicAuthUnsupported.ValueBoolPointer(),
		Scopes:                     conv.ForEachSliceItem(m.Scopes, func(v types.String) string { return v.ValueString() }),
		HostedDomains:              conv.ForEachSliceItem(m.HostedDomains, func(v types.String) string { return v.ValueString() }),
		RootCAs:                    conv.ForEachSliceItem(m.RootCAs, func(v types.String) string { return v.ValueString() }),
		InsecureSkipEmailVerified:  m.InsecureSkipEmailVerified.ValueBool(),
		InsecureEnableGroups:       m.InsecureEnableGroups.ValueBool(),
		AllowedGroups:              conv.ForEachSliceItem(m.AllowedGroups, func(v types.String) string { return v.ValueString() }),
		ACRValues:                  conv.ForEachSliceItem(m.ACRValues, func(v types.String) string { return v.ValueString() }),
		InsecureSkipVerify:         m.InsecureSkipVerify.ValueBool(),
		GetUserInfo:                m.GetUserInfo.ValueBool(),
		UserIDKey:                  m.UserIDKey.ValueString(),
		UserNameKey:                m.UserNameKey.ValueString(),
		PromptType:                 m.PromptType.ValueStringPointer(),
		OverrideClaimMapping:       m.OverrideClaimMapping.ValueBool(),
		ClaimMapping:               m.ClaimMapping.ToObject(),
		ClaimMutations:             m.ClaimMutations.ToObject(),
	}
}

// oidcProviderDiscoveryOverride contains the options to override the discovered urls.
type oidcProviderDiscoveryOverride struct {
	// TokenURL provides a way to user overwrite the Token URL
	// from the .well-known/openid-configuration token_endpoint
	TokenURL types.String `tfsdk:"token_url"`

	// AuthURL provides a way to user overwrite the Auth URL
	// from the .well-known/openid-configuration authorization_endpoint
	AuthURL types.String `tfsdk:"auth_url"`

	// JWKSURL provides a way to user overwrite the JWKS URL
	// from the .well-known/openid-configuration jwks_uri
	JWKSURL types.String `tfsdk:"jwks_url"`
}

func (m *oidcProviderDiscoveryOverride) ToObject() authv1.OIDCProviderDiscoveryOverrides {
	if m == nil {
		return authv1.OIDCProviderDiscoveryOverrides{}
	}
	return authv1.OIDCProviderDiscoveryOverrides{
		TokenURL: m.TokenURL.ValueString(),
		AuthURL:  m.AuthURL.ValueString(),
		JWKSURL:  m.JWKSURL.ValueString(),
	}
}

func newOIDCProviderDiscoveryOverride(override authv1.OIDCProviderDiscoveryOverrides) *oidcProviderDiscoveryOverride {
	if override == (authv1.OIDCProviderDiscoveryOverrides{}) {
		return nil
	}
	return &oidcProviderDiscoveryOverride{
		TokenURL: types.StringValue(override.TokenURL),
		AuthURL:  types.StringValue(override.AuthURL),
		JWKSURL:  types.StringValue(override.JWKSURL),
	}
}

// oidcProviderClaimMapping contains the claim mapping options.
type oidcProviderClaimMapping struct {
	// Configurable key which contains the preferred username claims
	// Defaults to "preferred_username".
	PreferredUsernameKey types.String `tfsdk:"preferred_username"`

	// Configurable key which contains the email claims.
	// Defaults to "email".
	EmailKey types.String `tfsdk:"email"`

	// Configurable key which contains the groups claims
	// Defaults to "groups".
	GroupsKey types.String `tfsdk:"groups"`
}

func (m *oidcProviderClaimMapping) ToObject() authv1.OIDCProviderClaimMapping {
	if m == nil {
		return authv1.OIDCProviderClaimMapping{}
	}
	return authv1.OIDCProviderClaimMapping{
		PreferredUsernameKey: m.PreferredUsernameKey.ValueString(),
		EmailKey:             m.EmailKey.ValueString(),
		GroupsKey:            m.GroupsKey.ValueString(),
	}
}

func newOIDCClaimMapping(cm authv1.OIDCProviderClaimMapping) *oidcProviderClaimMapping {
	if cm == (authv1.OIDCProviderClaimMapping{}) {
		return nil
	}
	return &oidcProviderClaimMapping{
		PreferredUsernameKey: types.StringValue(cm.PreferredUsernameKey),
		EmailKey:             types.StringValue(cm.EmailKey),
		GroupsKey:            types.StringValue(cm.GroupsKey),
	}
}

// oidcProviderClaimMutations contains the claim mutations options.
type oidcProviderClaimMutations struct {
	// NewGroupFromClaims specifies how claims can be joined to create groups.
	NewGroupFromClaims []oidcProviderNewGroupFromClaims `tfsdk:"new_group_from_claims"`

	// FilterGroupClaims is a regex filter used to keep only the matching groups.
	FilterGroupClaims *oidcProviderFilterGroupClaims `tfsdk:"filter_group_claims"`
}

func (m *oidcProviderClaimMutations) ToObject() authv1.OIDCProviderClaimMutations {
	if m == nil || (len(m.NewGroupFromClaims) == 0 && m.FilterGroupClaims == nil) {
		return authv1.OIDCProviderClaimMutations{}
	}
	return authv1.OIDCProviderClaimMutations{
		NewGroupFromClaims: conv.ForEachSliceItem(m.NewGroupFromClaims, func(claim oidcProviderNewGroupFromClaims) authv1.OIDCProviderNewGroupFromClaims {
			return claim.ToObject()
		}),
		FilterGroupClaims: m.FilterGroupClaims.ToObject(),
	}
}

func newOIDCClaimMutations(cm authv1.OIDCProviderClaimMutations) *oidcProviderClaimMutations {
	var filter *oidcProviderFilterGroupClaims
	if cm.FilterGroupClaims.GroupsFilter != "" {
		filter = &oidcProviderFilterGroupClaims{GroupsFilter: types.StringValue(cm.FilterGroupClaims.GroupsFilter)}
	}
	return &oidcProviderClaimMutations{
		NewGroupFromClaims: conv.ForEachSliceItem(cm.NewGroupFromClaims, newOIDCProviderNewGroupFromClaims),
		FilterGroupClaims:  filter,
	}
}

// oidcProviderNewGroupFromClaims is a list of claims to join together to create a new group.
type oidcProviderNewGroupFromClaims struct {
	// Claims is a list of claim to join together
	Claims []types.String `tfsdk:"claims"`

	// Delimiter is the string used to separate the claims.
	Delimiter types.String `tfsdk:"delimiter"`

	// ClearDelimiter indicates if the Delimiter string should be removed from claim values.
	ClearDelimiter types.Bool `tfsdk:"clear_delimiter"`

	// Prefix is a types.String to place before the first claim.
	Prefix types.String `tfsdk:"prefix"`
}

func (m *oidcProviderNewGroupFromClaims) ToObject() authv1.OIDCProviderNewGroupFromClaims {
	if m == nil {
		return authv1.OIDCProviderNewGroupFromClaims{}
	}
	return authv1.OIDCProviderNewGroupFromClaims{
		Claims:         conv.ForEachSliceItem(m.Claims, func(s types.String) string { return s.ValueString() }),
		Delimiter:      m.Delimiter.ValueString(),
		ClearDelimiter: m.ClearDelimiter.ValueBool(),
		Prefix:         m.Prefix.ValueString(),
	}
}

func newOIDCProviderNewGroupFromClaims(claim authv1.OIDCProviderNewGroupFromClaims) oidcProviderNewGroupFromClaims {
	return oidcProviderNewGroupFromClaims{
		Claims:         conv.ForEachSliceItem(claim.Claims, types.StringValue),
		Delimiter:      types.StringValue(claim.Delimiter),
		ClearDelimiter: types.BoolValue(claim.ClearDelimiter),
		Prefix:         types.StringValue(claim.Prefix),
	}
}

// oidcProviderFilterGroupClaims is a regex filter used to keep only the matching groups.
// This is useful when the groups list is too large to fit within an HTTP header.
type oidcProviderFilterGroupClaims struct {
	// GroupsFilter is the regex filter used to keep only the matching groups.
	GroupsFilter types.String `tfsdk:"groups_filter"`
}

func (m *oidcProviderFilterGroupClaims) ToObject() authv1.OIDCProviderFilterGroupClaims {
	if m == nil {
		return authv1.OIDCProviderFilterGroupClaims{}
	}
	return authv1.OIDCProviderFilterGroupClaims{
		GroupsFilter: m.GroupsFilter.ValueString(),
	}
}
