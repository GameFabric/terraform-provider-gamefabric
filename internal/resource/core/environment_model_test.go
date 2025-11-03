package core

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvironmentModel(t *testing.T) {
	model := newEnvironmentModel(testEnvironmentObject)
	assert.Equal(t, testEnvironmentModel, model)
}

func TestEnvironmentModel_ToObject(t *testing.T) {
	obj := testEnvironmentModel.ToObject()
	assert.Equal(t, testEnvironmentObject, obj)
}

var (
	testEnvironmentObject = &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-environment",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "Test Environment Display Name",
			Description: "Test Environment Description",
		},
	}

	testEnvironmentModel = environmentModel{
		ID:   types.StringValue("test-environment"),
		Name: types.StringValue("test-environment"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		DisplayName: types.StringValue("Test Environment Display Name"),
		Description: types.StringValue("Test Environment Description"),
	}
)
