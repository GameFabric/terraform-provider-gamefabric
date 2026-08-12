package validators

import (
	"context"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const maxGameNameLength = 20

// gameNameRegexp matches lowercase letters, numbers, and hyphens only.
//
// This mirrors GCAP's admission validation for spec.game.name (STS-2812,
// https://github.com/GameFabric/gcap/pull/441), giving plan-time feedback instead of
// waiting for an API round trip.
var gameNameRegexp = regexp.MustCompile(`^[a-z0-9-]*$`)

// GameNameValidator is a custom validator that checks if a string is a valid game name.
type GameNameValidator struct{}

// Description provides a description of the validator.
func (v GameNameValidator) Description(context.Context) string {
	return fmt.Sprintf("Validates that the value contains only lowercase letters, numbers, and hyphens, "+
		"with a maximum length of %d characters.", maxGameNameLength)
}

// MarkdownDescription provides a markdown description of the validator.
func (v GameNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString checks that the provided string is a valid game name.
func (v GameNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	switch {
	case value == "":
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid game name",
			"game_name is required",
		))
	case len(value) > maxGameNameLength:
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid game name length",
			fmt.Sprintf("game_name must be no more than %d characters, got %d", maxGameNameLength, len(value)),
		))
	case !gameNameRegexp.MatchString(value):
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid game name",
			"game_name must contain only lowercase letters, numbers, and hyphens",
		))
	}
}

// TokenServiceAllowedPlatformNames are the platform names accepted by the Token Service,
// sorted, matching sts-svc-token-service-provisioner's AllowedPlatformNames(). The "eos"
// sentinel used internally by GCAP for EOS mode is intentionally excluded: the Terraform
// resource exposes EOS configuration through the dedicated eos attribute instead.
var TokenServiceAllowedPlatformNames = []string{
	"android",
	"ios",
	"pc",
	"playstation",
	"ps4",
	"switch",
	"xbox",
}

// TokenServiceEOSPlatformName is the platform name GCAP reserves for EOS mode. The
// provider writes it transparently (STS-2888 bridge) and hides it again on read; users
// never configure it.
const TokenServiceEOSPlatformName = "eos"

// TokenServiceSigningAlgorithms are the signing algorithms accepted for game client token keys.
var TokenServiceSigningAlgorithms = []string{
	"HS256", "HS384", "HS512",
	"RS256", "RS384", "RS512",
	"PS256", "PS384", "PS512",
	"ES256", "ES384", "ES512",
}

var symmetricSigningAlgorithms = []string{"HS256", "HS384", "HS512"}

// pemBlockTypesByAlgorithm are the PEM block types accepted as a public key for each
// asymmetric signing algorithm. Private keys and certificates are intentionally excluded,
// since only public keys are expected here. Mirrors GCAP's admission.allowedPEMBlockTypes.
var pemBlockTypesByAlgorithm = map[string][]string{
	"RS256": {"PUBLIC KEY", "RSA PUBLIC KEY"},
	"RS384": {"PUBLIC KEY", "RSA PUBLIC KEY"},
	"RS512": {"PUBLIC KEY", "RSA PUBLIC KEY"},
	"PS256": {"PUBLIC KEY", "RSA PUBLIC KEY"},
	"PS384": {"PUBLIC KEY", "RSA PUBLIC KEY"},
	"PS512": {"PUBLIC KEY", "RSA PUBLIC KEY"},
	"ES256": {"PUBLIC KEY"},
	"ES384": {"PUBLIC KEY"},
	"ES512": {"PUBLIC KEY"},
}

// TokenServiceKeyMaterialValidator validates that a game client token key's key material
// matches the format expected for its signing algorithm: base64 for HMAC algorithms
// (HS256/384/512), or PEM-encoded public key for RSA/ECDSA algorithms. This mirrors GCAP's
// admission.validateKeyMaterial and requires both the "key" and "signing_algorithm"
// sibling attributes, so it is applied as an object-level validator on the
// game_client_token_keys NestedAttributeObject.
type TokenServiceKeyMaterialValidator struct{}

// Description provides a description of the validator.
func (v TokenServiceKeyMaterialValidator) Description(context.Context) string {
	return "Validates that key material matches the format expected for the given signing algorithm: " +
		"base64 for HMAC algorithms, PEM-encoded public key for RSA/ECDSA algorithms."
}

// MarkdownDescription provides a markdown description of the validator.
func (v TokenServiceKeyMaterialValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateObject checks that the key material matches the format expected for the
// configured signing algorithm.
func (v TokenServiceKeyMaterialValidator) ValidateObject(
	_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	key, alg, ok := keyAndAlgorithm(attrs)
	if !ok {
		return
	}

	if errMsg := validateKeyMaterial(key, alg); errMsg != "" {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path.AtName("key"),
			"Invalid key material",
			errMsg,
		))
	}
}

// keyAndAlgorithm extracts the key and signing_algorithm sibling values, returning ok=false
// if either is missing, null, unknown, or empty (in which case there is nothing to validate
// yet; other validators cover those cases).
func keyAndAlgorithm(attrs map[string]attr.Value) (key, alg string, ok bool) {
	keyVal, ok := attrs["key"].(basetypes.StringValue)
	if !ok || keyVal.IsNull() || keyVal.IsUnknown() || keyVal.ValueString() == "" {
		return "", "", false
	}
	algVal, ok := attrs["signing_algorithm"].(basetypes.StringValue)
	if !ok || algVal.IsNull() || algVal.IsUnknown() || algVal.ValueString() == "" {
		return "", "", false
	}
	return keyVal.ValueString(), algVal.ValueString(), true
}

// validateKeyMaterial returns a non-empty error message if key does not match the format
// expected for alg: base64 for HMAC algorithms, or a PEM-encoded public key of the
// appropriate type for RSA/ECDSA algorithms. Returns an empty string if alg is not a
// recognized signing algorithm, deferring to the signing_algorithm OneOf validator.
func validateKeyMaterial(key, alg string) string {
	if slices.Contains(symmetricSigningAlgorithms, alg) {
		if _, err := base64.StdEncoding.DecodeString(key); err != nil {
			return fmt.Sprintf("key is not valid base64 for signing algorithm %s: %s", alg, err)
		}
		return ""
	}

	allowedTypes, ok := pemBlockTypesByAlgorithm[alg]
	if !ok {
		// Unknown algorithm: the signing_algorithm OneOf validator already reports this.
		return ""
	}

	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return "key is not a valid PEM-encoded key"
	}
	if !slices.Contains(allowedTypes, block.Type) {
		return fmt.Sprintf("PEM block of type %q is not a valid public key for algorithm %s, expected one of %v",
			block.Type, alg, allowedTypes)
	}
	return ""
}
