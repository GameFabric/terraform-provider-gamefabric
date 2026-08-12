package provisioning

import (
	"encoding/json"
	"strings"
	"time"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
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

func newTokenServiceModel(obj *provisioningv1beta1.TokenService) tokenServiceModel {
	hasEOS := len(obj.Spec.EOS) > 0
	return tokenServiceModel{
		ID:               types.StringValue(obj.Name),
		Name:             types.StringValue(obj.Name),
		DevelopmentMode:  types.BoolValue(obj.Spec.Environment == provisioningv1beta1.TokenServiceEnvDev),
		Labels:           conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:      conv.ForEachMapItem(obj.Annotations, types.StringValue),
		GameName:         types.StringValue(obj.Spec.Game.Name),
		Platforms:        newTokenServicePlatformModels(obj.Spec.Game.Platforms, hasEOS),
		EOS:              newTokenServiceEOSModel(obj.Spec.EOS),
		JWKS:             newTokenServiceJWKSModel(obj.Spec.JWKS),
		State:            types.StringValue(string(obj.Status.State)),
		Hostname:         conv.OptionalFunc(obj.Status.Hostname, types.StringValue, types.StringNull),
		PlatformKey:      FirstPlatformKey(obj.Status.PlatformKeys),
		Reason:           conv.OptionalFunc(obj.Status.Reason, types.StringValue, types.StringNull),
		StateLastChanged: newStateLastChanged(obj.Status.StateLastChanged),
	}
}

func (m tokenServiceModel) ToObject() *provisioningv1beta1.TokenService {
	return &provisioningv1beta1.TokenService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(v types.String) string { return v.ValueString() }),
		},
		Spec: provisioningv1beta1.TokenServiceSpec{
			Environment: tokenServiceEnvFromDevelopmentMode(m.DevelopmentMode.ValueBool()),
			EOS:         tokenServiceEOSToObject(m.EOS),
			Game: provisioningv1beta1.TokenServiceGameSpec{
				Name:      m.GameName.ValueString(),
				Platforms: tokenServicePlatformsToObject(m.Platforms, len(m.EOS) > 0),
			},
			JWKS: tokenServiceJWKSToObject(m.JWKS),
		},
	}
}

// tokenServiceEnvFromDevelopmentMode maps the development_mode boolean to the environment
// string the GameFabric API currently expects: true -> dev, false -> prod.
func tokenServiceEnvFromDevelopmentMode(developmentMode bool) provisioningv1beta1.TokenServiceEnv {
	if developmentMode {
		return provisioningv1beta1.TokenServiceEnvDev
	}
	return provisioningv1beta1.TokenServiceEnvProd
}

// newTokenServicePlatformModels converts the backend's platforms map to the model's
// platforms map, hiding the "eos" platform entry when hasEOS is true (STS-2888 bridge).
func newTokenServicePlatformModels(
	platforms map[string]provisioningv1beta1.TokenServicePlatformSpec, hasEOS bool,
) map[string]tokenServicePlatformModel {
	out := make(map[string]tokenServicePlatformModel, len(platforms))
	for name, platform := range platforms {
		if hasEOS && name == validators.TokenServiceEOSPlatformName {
			continue
		}
		out[name] = tokenServicePlatformModel{
			GameClientTokenKeys: conv.ForEachSliceItem(platform.GameClientTokenKeys, newTokenServiceKeyModel),
		}
	}
	return out
}

// tokenServicePlatformsToObject converts the model's platforms map to the backend's
// platforms map, adding the "eos" platform entry when hasEOS is true (STS-2888 bridge).
func tokenServicePlatformsToObject(
	platforms map[string]tokenServicePlatformModel, hasEOS bool,
) map[string]provisioningv1beta1.TokenServicePlatformSpec {
	out := make(map[string]provisioningv1beta1.TokenServicePlatformSpec, len(platforms))
	for name, platform := range platforms {
		out[name] = provisioningv1beta1.TokenServicePlatformSpec{
			GameClientTokenKeys: conv.ForEachSliceItem(platform.GameClientTokenKeys, tokenServiceKeyToObject),
		}
	}
	if hasEOS {
		out[validators.TokenServiceEOSPlatformName] = provisioningv1beta1.TokenServicePlatformSpec{}
	}
	return out
}

func newTokenServiceKeyModel(key provisioningv1beta1.TokenServiceKeySpec) tokenServiceKeyModel {
	return tokenServiceKeyModel{
		Key:              types.StringValue(key.Key),
		SigningAlgorithm: types.StringValue(string(key.SigningAlgorithm)),
	}
}

func tokenServiceKeyToObject(key tokenServiceKeyModel) provisioningv1beta1.TokenServiceKeySpec {
	return provisioningv1beta1.TokenServiceKeySpec{
		Key:              key.Key.ValueString(),
		SigningAlgorithm: provisioningv1beta1.TokenServiceSigningAlgorithm(key.SigningAlgorithm.ValueString()),
	}
}

// newTokenServiceEOSModel converts the backend's EOS list to the model's EOS list.
// Returns nil if no EOS configuration is present.
func newTokenServiceEOSModel(eos []provisioningv1beta1.TokenServiceEOSSpec) []tokenServiceEOSModel {
	if len(eos) == 0 {
		return nil
	}
	out := make([]tokenServiceEOSModel, len(eos))
	for i, e := range eos {
		out[i] = tokenServiceEOSModel{
			ClientID:     types.StringValue(e.ClientID),
			TokenTypes:   conv.ForEachSliceItem(e.TokenTypes, func(t provisioningv1beta1.TokenServiceEOSTokenType) types.String { return types.StringValue(string(t)) }),
			DeploymentID: conv.OptionalFunc(e.DeploymentID, types.StringValue, types.StringNull),
			ProductID:    conv.OptionalFunc(e.ProductID, types.StringValue, types.StringNull),
			SandboxID:    conv.OptionalFunc(e.SandboxID, types.StringValue, types.StringNull),
		}
	}
	return out
}

// tokenServiceEOSToObject converts the model's EOS list to the backend's EOS list,
// or nil if the eos attribute is empty.
func tokenServiceEOSToObject(m []tokenServiceEOSModel) []provisioningv1beta1.TokenServiceEOSSpec {
	if len(m) == 0 {
		return nil
	}
	out := make([]provisioningv1beta1.TokenServiceEOSSpec, len(m))
	for i, e := range m {
		out[i] = provisioningv1beta1.TokenServiceEOSSpec{
			ClientID:     e.ClientID.ValueString(),
			DeploymentID: e.DeploymentID.ValueString(),
			ProductID:    e.ProductID.ValueString(),
			SandboxID:    e.SandboxID.ValueString(),
			TokenTypes: conv.ForEachSliceItem(e.TokenTypes, func(v types.String) provisioningv1beta1.TokenServiceEOSTokenType {
				return provisioningv1beta1.TokenServiceEOSTokenType(v.ValueString())
			}),
		}
	}
	return out
}

func newTokenServiceJWKSModel(jwks *provisioningv1beta1.TokenServiceJWKSSpec) *tokenServiceJWKSModel {
	if jwks == nil {
		return nil
	}
	return &tokenServiceJWKSModel{URL: types.StringValue(jwks.URL)}
}

func tokenServiceJWKSToObject(m *tokenServiceJWKSModel) *provisioningv1beta1.TokenServiceJWKSSpec {
	if m == nil {
		return nil
	}
	return &provisioningv1beta1.TokenServiceJWKSSpec{URL: m.URL.ValueString()}
}

func newStateLastChanged(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}

// FirstPlatformKey extracts a single "platform key" from the legacy platformKeys status
// field, which is a JSON object mapping platform name to a list of keys, e.g.
// `{"pc": ["key1", "key2"], "xbox": ["key3"]}`.
//
// GCAP does not yet expose a singular platform key field (see STS-2739/STS-2740), so, per
// the resource spec, this mimics the GameFabric webui's extraction logic
// (TokenServices.vue's getPlatformKeyValue): take the first platform in the JSON object
// (in original key order) and the first key in its list.
//
// Go's map[string]T unmarshaling does not preserve key order, so this walks the raw JSON
// tokens instead to match the frontend's Object.values(parsed)[0][0] behavior, which relies
// on JavaScript's insertion-order iteration of string object keys.
//
// Returns a null value if platformKeysJSON is empty, syntactically malformed (validated in
// full, not just up to the first extracted value), or has no keys, matching the webui's
// fallback to an empty value on any parse error.
func FirstPlatformKey(platformKeysJSON string) types.String {
	if platformKeysJSON == "" || !json.Valid([]byte(platformKeysJSON)) {
		return types.StringNull()
	}

	dec := json.NewDecoder(strings.NewReader(platformKeysJSON))

	tok, err := dec.Token()
	if err != nil {
		return types.StringNull()
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return types.StringNull()
	}

	if !dec.More() {
		// Empty object.
		return types.StringNull()
	}

	// The first object key, in original JSON order.
	if _, err = dec.Token(); err != nil {
		return types.StringNull()
	}

	var keys []string
	if err = dec.Decode(&keys); err != nil {
		return types.StringNull()
	}
	if len(keys) == 0 {
		return types.StringNull()
	}

	return types.StringValue(keys[0])
}
