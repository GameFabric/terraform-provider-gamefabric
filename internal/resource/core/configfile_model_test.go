package core

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigModel(t *testing.T) {
	model := newConfigModel(testConfigFileObject)
	assert.Equal(t, testConfigFileModel, model)
}

func TestConfigModel_ToObject(t *testing.T) {
	obj := testConfigFileModel.ToObject()
	assert.Equal(t, testConfigFileObject, obj)
}

var (
	testConfigFileObject = &corev1.ConfigFile{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-configfile",
			Environment: "test-environment",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Description: "ConfigFile Description",
		Data:        "config file data content",
	}

	testConfigFileModel = configFileModel{
		ID:          types.StringValue("test-environment/test-configfile"),
		Name:        types.StringValue("test-configfile"),
		Environment: types.StringValue("test-environment"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		Description: types.StringValue("ConfigFile Description"),
		Data:        types.StringValue("config file data content"),
	}
)
