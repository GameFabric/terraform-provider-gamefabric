package protection

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewGatewayPolicyModel(t *testing.T) {
	model := newGatewayPolicyModel(testGatewayPolicyObject)
	assert.Equal(t, testGatewayPolicyModel, model)
}

func TestGatewayPolicyModel_ToObject(t *testing.T) {
	obj := testGatewayPolicyModel.ToObject()
	assert.Equal(t, testGatewayPolicyObject, obj)
}

var (
	testGatewayPolicyObject = &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gateway-policy",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: protectionv1.GatewayPolicySpec{
			DisplayName: "Test Gateway Policy",
			Description: "Test Gateway Policy Description",
			DestinationCIDRs: []string{
				"10.0.0.0/8",
				"192.168.1.0/24",
			},
		},
	}

	testGatewayPolicyModel = gatewayPolicyModel{
		ID:   types.StringValue("test-gateway-policy"),
		Name: types.StringValue("test-gateway-policy"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		DisplayName: types.StringValue("Test Gateway Policy"),
		Description: types.StringValue("Test Gateway Policy Description"),
		DestinationCIDRs: []types.String{
			types.StringValue("10.0.0.0/8"),
			types.StringValue("192.168.1.0/24"),
		},
	}
)
