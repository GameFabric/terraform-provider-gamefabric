package core

import (
	"testing"

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
		Name:        "test-config",
		Environment: "test-environment",
		Data:        "test-data-content",
	}

	testConfigFileModel = configFileModel{
		Name:        types.StringValue("test-config"),
		Environment: types.StringValue("test-environment"),
		Data:        types.StringValue("test-data-content"),
	}
)
