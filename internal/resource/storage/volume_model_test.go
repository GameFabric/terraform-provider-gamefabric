package storage

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewVolumeModel(t *testing.T) {
	model := newVolumeModel(testVolumeObject)

	assert.Equal(t, testVolumeModel, model)
}

func TestVolumeModel_ToObject(t *testing.T) {
	obj := testVolumeModel.ToObject()

	assert.Equal(t, testVolumeObject, obj)
}

var (
	testVolumeObject = &storagev1beta1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-volume",
			Environment: "dflt",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: storagev1beta1.VolumeSpec{
			VolumeStoreName: "test-volume-store",
			Capacity:        resource.MustParse("1G"),
		},
	}

	testVolumeModel = volumeModel{
		ID:          types.StringValue("dflt/test-volume"),
		Name:        types.StringValue("test-volume"),
		Environment: types.StringValue("dflt"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		VolumeStore: types.StringValue("test-volume-store"),
		Capacity:    types.StringValue("1G"),
	}
)
