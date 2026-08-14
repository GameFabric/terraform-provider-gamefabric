package provisioning

import (
	"time"

	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	resourceprovisioning "github.com/gamefabric/terraform-provider-gamefabric/internal/resource/provisioning"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tokenServiceModel struct {
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
	return tokenServiceModel{
		Name:             types.StringValue(obj.Name),
		DevelopmentMode:  types.BoolValue(obj.Spec.Environment == provisioningv1beta1.TokenServiceEnvDev),
		Labels:           conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:      conv.ForEachMapItem(obj.Annotations, types.StringValue),
		GameName:         types.StringValue(obj.Spec.Game.Name),
		Platforms:        newTokenServicePlatformModels(obj.Spec.Game.Platforms, len(obj.Spec.EOS) > 0),
		EOS:              newTokenServiceEOSModel(obj.Spec.EOS),
		JWKS:             newTokenServiceJWKSModel(obj.Spec.JWKS),
		State:            types.StringValue(string(obj.Status.State)),
		Hostname:         conv.OptionalFunc(obj.Status.Hostname, types.StringValue, types.StringNull),
		PlatformKey:      resourceprovisioning.FirstPlatformKey(obj.Status.PlatformKeys),
		Reason:           conv.OptionalFunc(obj.Status.Reason, types.StringValue, types.StringNull),
		StateLastChanged: newStateLastChanged(obj.Status.StateLastChanged),
	}
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

func newTokenServiceKeyModel(key provisioningv1beta1.TokenServiceKeySpec) tokenServiceKeyModel {
	return tokenServiceKeyModel{
		Key:              types.StringValue(key.Key),
		SigningAlgorithm: types.StringValue(string(key.SigningAlgorithm)),
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
			ClientID: types.StringValue(e.ClientID),
			TokenTypes: conv.ForEachSliceItem(e.TokenTypes, func(t provisioningv1beta1.TokenServiceEOSTokenType) types.String {
				return types.StringValue(string(t))
			}),
			DeploymentID: conv.OptionalFunc(e.DeploymentID, types.StringValue, types.StringNull),
			ProductID:    conv.OptionalFunc(e.ProductID, types.StringValue, types.StringNull),
			SandboxID:    conv.OptionalFunc(e.SandboxID, types.StringValue, types.StringNull),
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

func newStateLastChanged(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
