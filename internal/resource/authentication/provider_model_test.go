package authentication

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authenticationv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/hamba/pkg/v2/ptr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewProviderModel(t *testing.T) {
	model := newProviderModel(testProviderObject)
	assert.Equal(t, testProviderModel, model)
}

func TestProviderModel_ToObject(t *testing.T) {
	obj := testProviderModel.ToObject()
	assert.Equal(t, testProviderObject, obj)
}

var (
	testProviderObject = &authenticationv1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-provider",
		},
		Spec: authenticationv1.ProviderSpec{
			DisplayName: "Test Provider",
			OIDC: &authenticationv1.OIDCProvider{
				Issuer:       "issuer",
				ClientID:     "client",
				ClientSecret: "secr3t",
				RedirectURI:  "https://example.org/redirecturi",
				ProviderDiscoveryOverrides: authenticationv1.OIDCProviderDiscoveryOverrides{
					TokenURL: "https://example.org/token",
					AuthURL:  "https://example.org/auth",
					JWKSURL:  "https://example.org/jwks",
				},
				BasicAuthUnsupported:      ptr.Of(true),
				Scopes:                    []string{"openid", "email", "profile"},
				HostedDomains:             []string{"example.org", "example.com"},
				RootCAs:                   []string{"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"},
				InsecureSkipEmailVerified: false,
				InsecureEnableGroups:      true,
				AllowedGroups:             []string{"admins", "devs"},
				ACRValues:                 []string{"urn:mace:incommon:iap:silver"},
				InsecureSkipVerify:        true,
				GetUserInfo:               true,
				UserIDKey:                 "userid",
				UserNameKey:               "username",
				PromptType:                ptr.Of("login"),
				OverrideClaimMapping:      true,
				ClaimMapping: authenticationv1.OIDCProviderClaimMapping{
					PreferredUsernameKey: "preferred_username",
					EmailKey:             "email",
					GroupsKey:            "groups",
				},
				ClaimMutations: authenticationv1.OIDCProviderClaimMutations{
					NewGroupFromClaims: []authenticationv1.OIDCProviderNewGroupFromClaims{
						{
							Claims:         []string{"claim1", "claim2"},
							Delimiter:      "-",
							ClearDelimiter: true,
							Prefix:         "role_",
						},
					},
					FilterGroupClaims: authenticationv1.OIDCProviderFilterGroupClaims{
						GroupsFilter: "^dev-.*$",
					},
				},
			},
		},
	}

	testProviderModel = providerModel{
		ID:          types.StringValue("test-provider"),
		Name:        types.StringValue("test-provider"),
		DisplayName: types.StringValue("Test Provider"),
		OIDC: &oidcProvider{
			Issuer:       types.StringValue("issuer"),
			ClientID:     types.StringValue("client"),
			ClientSecret: types.StringValue("secr3t"),
			RedirectURI:  types.StringValue("https://example.org/redirecturi"),
			ProviderDiscoveryOverrides: &oidcProviderDiscoveryOverride{
				TokenURL: types.StringValue("https://example.org/token"),
				AuthURL:  types.StringValue("https://example.org/auth"),
				JWKSURL:  types.StringValue("https://example.org/jwks"),
			},
			BasicAuthUnsupported: types.BoolValue(true),
			Scopes: []types.String{
				types.StringValue("openid"),
				types.StringValue("email"),
				types.StringValue("profile"),
			},
			HostedDomains: []types.String{
				types.StringValue("example.org"),
				types.StringValue("example.com"),
			},
			RootCAs: []types.String{
				types.StringValue("-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"),
			},
			InsecureSkipEmailVerified: types.BoolValue(false),
			InsecureEnableGroups:      types.BoolValue(true),
			AllowedGroups: []types.String{
				types.StringValue("admins"),
				types.StringValue("devs"),
			},
			ACRValues: []types.String{
				types.StringValue("urn:mace:incommon:iap:silver"),
			},
			InsecureSkipVerify:   types.BoolValue(true),
			GetUserInfo:          types.BoolValue(true),
			UserIDKey:            types.StringValue("userid"),
			UserNameKey:          types.StringValue("username"),
			PromptType:           types.StringValue("login"),
			OverrideClaimMapping: types.BoolValue(true),
			ClaimMapping: &oidcProviderClaimMapping{
				PreferredUsernameKey: types.StringValue("preferred_username"),
				EmailKey:             types.StringValue("email"),
				GroupsKey:            types.StringValue("groups"),
			},
			ClaimMutations: &oidcProviderClaimMutations{
				NewGroupFromClaims: []oidcProviderNewGroupFromClaims{
					{
						Claims: []types.String{
							types.StringValue("claim1"),
							types.StringValue("claim2"),
						},
						Delimiter:      types.StringValue("-"),
						ClearDelimiter: types.BoolValue(true),
						Prefix:         types.StringValue("role_"),
					},
				},
				FilterGroupClaims: &oidcProviderFilterGroupClaims{
					GroupsFilter: types.StringValue("^dev-.*$"),
				},
			},
		},
	}
)
