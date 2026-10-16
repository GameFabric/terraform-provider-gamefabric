package provisioning

import (
	"time"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta3 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta3"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tokenServiceModel struct {
	ID               types.String                         `tfsdk:"id"`
	Name             types.String                         `tfsdk:"name"`
	DevelopmentMode  types.Bool                           `tfsdk:"development_mode"`
	Labels           map[string]types.String              `tfsdk:"labels"`
	Annotations      map[string]types.String              `tfsdk:"annotations"`
	GameName         types.String                         `tfsdk:"game_name"`
	Platforms        map[string]tokenServicePlatformModel `tfsdk:"platforms"`
	EOS              []tokenServiceEOSModel               `tfsdk:"eos"`
	JWKS             *tokenServiceJWKSModel               `tfsdk:"jwks"`
	State            types.String                         `tfsdk:"state"`
	Hostname         types.String                         `tfsdk:"hostname"`
	PlatformKey      types.String                         `tfsdk:"platform_key"`
	Reason           types.String                         `tfsdk:"reason"`
	StateLastChanged types.String                         `tfsdk:"state_last_changed"`
}

// tokenServicePlatformModel is the model for a single platform's verification configuration.
type tokenServicePlatformModel struct {
	GameClientTokenKeys []tokenServiceKeyModel `tfsdk:"game_client_token_keys"`
}

// tokenServiceKeyModel is the model for a game client token signing key.
type tokenServiceKeyModel struct {
	Key              types.String `tfsdk:"key"`
	SigningAlgorithm types.String `tfsdk:"signing_algorithm"`
}

// tokenServiceEOSModel is the model for a single Epic Online Services (EOS) auth mode entry.
type tokenServiceEOSModel struct {
	ClientID     types.String   `tfsdk:"client_id"`
	TokenTypes   []types.String `tfsdk:"token_types"`
	DeploymentID types.String   `tfsdk:"deployment_id"`
	ProductID    types.String   `tfsdk:"product_id"`
	SandboxID    types.String   `tfsdk:"sandbox_id"`
}

// tokenServiceJWKSModel is the model for the JWKS auth mode.
type tokenServiceJWKSModel struct {
	URL types.String `tfsdk:"url"`
}

func newTokenServiceModel(obj *provisioningv1beta3.TokenService) tokenServiceModel {
	return tokenServiceModel{
		ID:               types.StringValue(obj.Name),
		Name:             types.StringValue(obj.Name),
		DevelopmentMode:  types.BoolValue(obj.Spec.DevelopmentMode),
		Labels:           conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:      conv.ForEachMapItem(obj.Annotations, types.StringValue),
		GameName:         types.StringValue(obj.Spec.Game.Name),
		Platforms:        newTokenServicePlatformModels(obj.Spec.Game.Platforms),
		EOS:              newTokenServiceEOSModel(obj.Spec.EOS),
		JWKS:             newTokenServiceJWKSModel(obj.Spec.JWKS),
		State:            types.StringValue(string(obj.Status.State)),
		Hostname:         conv.OptionalFunc(obj.Status.Hostname, types.StringValue, types.StringNull),
		PlatformKey:      conv.OptionalFunc(obj.Status.PlatformKey, types.StringValue, types.StringNull),
		Reason:           conv.OptionalFunc(obj.Status.Reason, types.StringValue, types.StringNull),
		StateLastChanged: newStateLastChanged(obj.Status.StateLastChanged),
	}
}

func (m tokenServiceModel) ToObject() *provisioningv1beta3.TokenService {
	return &provisioningv1beta3.TokenService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(v types.String) string { return v.ValueString() }),
		},
		Spec: provisioningv1beta3.TokenServiceSpec{
			DevelopmentMode: m.DevelopmentMode.ValueBool(),
			EOS:             tokenServiceEOSToObject(m.EOS),
			Game: provisioningv1beta3.TokenServiceGameSpec{
				Name:      m.GameName.ValueString(),
				Platforms: tokenServicePlatformsToObject(m.Platforms),
			},
			JWKS: tokenServiceJWKSToObject(m.JWKS),
		},
	}
}

func newTokenServicePlatformModels(platforms map[string]provisioningv1beta3.TokenServicePlatformSpec) map[string]tokenServicePlatformModel {
	out := make(map[string]tokenServicePlatformModel, len(platforms))
	for name, platform := range platforms {
		out[name] = tokenServicePlatformModel{
			GameClientTokenKeys: conv.ForEachSliceItem(platform.GameClientTokenKeys, newTokenServiceKeyModel),
		}
	}
	return out
}

func tokenServicePlatformsToObject(platforms map[string]tokenServicePlatformModel) map[string]provisioningv1beta3.TokenServicePlatformSpec {
	out := make(map[string]provisioningv1beta3.TokenServicePlatformSpec, len(platforms))
	for name, platform := range platforms {
		out[name] = provisioningv1beta3.TokenServicePlatformSpec{
			GameClientTokenKeys: conv.ForEachSliceItem(platform.GameClientTokenKeys, tokenServiceKeyToObject),
		}
	}
	return out
}

func newTokenServiceKeyModel(key provisioningv1beta3.TokenServiceKeySpec) tokenServiceKeyModel {
	return tokenServiceKeyModel{
		Key:              types.StringValue(key.Key),
		SigningAlgorithm: types.StringValue(string(key.SigningAlgorithm)),
	}
}

func tokenServiceKeyToObject(key tokenServiceKeyModel) provisioningv1beta3.TokenServiceKeySpec {
	return provisioningv1beta3.TokenServiceKeySpec{
		Key:              key.Key.ValueString(),
		SigningAlgorithm: provisioningv1beta3.TokenServiceSigningAlgorithm(key.SigningAlgorithm.ValueString()),
	}
}

// newTokenServiceEOSModel converts the backend's EOS list to the model's EOS list.
// Returns nil if no EOS configuration is present.
func newTokenServiceEOSModel(eos []provisioningv1beta3.TokenServiceEOSSpec) []tokenServiceEOSModel {
	if len(eos) == 0 {
		return nil
	}
	out := make([]tokenServiceEOSModel, len(eos))
	for i, e := range eos {
		out[i] = tokenServiceEOSModel{
			ClientID:     types.StringValue(e.ClientID),
			TokenTypes:   conv.ForEachSliceItem(e.TokenTypes, func(t provisioningv1beta3.TokenServiceEOSTokenType) types.String { return types.StringValue(string(t)) }),
			DeploymentID: conv.OptionalFunc(e.DeploymentID, types.StringValue, types.StringNull),
			ProductID:    conv.OptionalFunc(e.ProductID, types.StringValue, types.StringNull),
			SandboxID:    conv.OptionalFunc(e.SandboxID, types.StringValue, types.StringNull),
		}
	}
	return out
}

// tokenServiceEOSToObject converts the model's EOS list to the backend's EOS list,
// or nil if the eos attribute is empty.
func tokenServiceEOSToObject(m []tokenServiceEOSModel) []provisioningv1beta3.TokenServiceEOSSpec {
	if len(m) == 0 {
		return nil
	}
	out := make([]provisioningv1beta3.TokenServiceEOSSpec, len(m))
	for i, e := range m {
		out[i] = provisioningv1beta3.TokenServiceEOSSpec{
			ClientID:     e.ClientID.ValueString(),
			DeploymentID: e.DeploymentID.ValueString(),
			ProductID:    e.ProductID.ValueString(),
			SandboxID:    e.SandboxID.ValueString(),
			TokenTypes: conv.ForEachSliceItem(e.TokenTypes, func(v types.String) provisioningv1beta3.TokenServiceEOSTokenType {
				return provisioningv1beta3.TokenServiceEOSTokenType(v.ValueString())
			}),
		}
	}
	return out
}

func newTokenServiceJWKSModel(jwks *provisioningv1beta3.TokenServiceJWKSSpec) *tokenServiceJWKSModel {
	if jwks == nil {
		return nil
	}
	return &tokenServiceJWKSModel{URL: types.StringValue(jwks.URL)}
}

func tokenServiceJWKSToObject(m *tokenServiceJWKSModel) *provisioningv1beta3.TokenServiceJWKSSpec {
	if m == nil {
		return nil
	}
	return &provisioningv1beta3.TokenServiceJWKSSpec{URL: m.URL.ValueString()}
}

func newStateLastChanged(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
