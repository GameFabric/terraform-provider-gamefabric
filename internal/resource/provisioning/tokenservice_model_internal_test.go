package provisioning

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta3 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta3"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTokenServiceModel_ToObjectDevelopmentMode(t *testing.T) {
	tests := []struct {
		name string
		in   types.Bool
		want bool
	}{
		{
			name: "true",
			in:   types.BoolValue(true),
			want: true,
		},
		{
			name: "false",
			in:   types.BoolValue(false),
			want: false,
		},
		{
			name: "null",
			in:   types.BoolNull(),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj := testTokenServiceModelWithDevelopmentMode(test.in).ToObject()
			assert.Equal(t, test.want, obj.Spec.DevelopmentMode)
		})
	}
}

func TestTokenServiceModel_NewFromObjectDevelopmentMode(t *testing.T) {
	tests := []struct {
		name string
		in   bool
		want types.Bool
	}{
		{
			name: "true",
			in:   true,
			want: types.BoolValue(true),
		},
		{
			name: "false",
			in:   false,
			want: types.BoolValue(false),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := newTokenServiceModel(testTokenServiceObjectWithDevelopmentMode(test.in))
			assert.Equal(t, test.want, model.DevelopmentMode)
		})
	}
}

func TestTokenServiceModel_NewFromObjectPlatformKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want types.String
	}{
		{
			name: "status platform key",
			in:   "key1",
			want: types.StringValue("key1"),
		},
		{
			name: "empty status platform key",
			in:   "",
			want: types.StringNull(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj := testTokenServiceObjectWithDevelopmentMode(true)
			obj.Status.PlatformKey = test.in

			model := newTokenServiceModel(obj)
			assert.Equal(t, test.want, model.PlatformKey)
		})
	}
}

func testTokenServiceModelWithDevelopmentMode(devMode types.Bool) tokenServiceModel {
	return tokenServiceModel{
		ID:              types.StringValue("my-service"),
		Name:            types.StringValue("my-service"),
		DevelopmentMode: devMode,
		GameName:        types.StringValue("my-game"),
		Platforms:       map[string]tokenServicePlatformModel{"pc": {GameClientTokenKeys: testTokenServiceKeyModels()}},
	}
}

func testTokenServiceObjectWithDevelopmentMode(devMode bool) *provisioningv1beta3.TokenService {
	return &provisioningv1beta3.TokenService{
		ObjectMeta: metav1.ObjectMeta{Name: "my-service"},
		Spec: provisioningv1beta3.TokenServiceSpec{
			DevelopmentMode: devMode,
			Game: provisioningv1beta3.TokenServiceGameSpec{
				Name:      "my-game",
				Platforms: map[string]provisioningv1beta3.TokenServicePlatformSpec{"pc": {GameClientTokenKeys: testTokenServiceKeySpecs()}},
			},
		},
	}
}

func testTokenServiceKeySpecs() []provisioningv1beta3.TokenServiceKeySpec {
	return []provisioningv1beta3.TokenServiceKeySpec{
		{Key: "MDEyMzQ1Njc4OWFiY2RlZg==", SigningAlgorithm: provisioningv1beta3.TokenServiceSigningAlgorithmHS256},
	}
}

func testTokenServiceKeyModels() []tokenServiceKeyModel {
	return []tokenServiceKeyModel{
		{Key: types.StringValue("MDEyMzQ1Njc4OWFiY2RlZg=="), SigningAlgorithm: types.StringValue("HS256")},
	}
}
