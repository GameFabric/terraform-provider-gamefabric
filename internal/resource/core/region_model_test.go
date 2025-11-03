package core

import (
	"testing"

	"agones.dev/agones/pkg/apis"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	kcorev1 "k8s.io/api/core/v1"
)

func TestNewRegionModel(t *testing.T) {
	model := newRegionModel(testRegionObject)
	assert.Equal(t, testRegionModel, model)
}

func TestRegionModel_ToObject(t *testing.T) {
	obj := testRegionModel.ToObject()
	assert.Equal(t, testRegionObject, obj)
}

var (
	testRegionObject = &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-region",
			Environment: "test-environment",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: corev1.RegionSpec{
			DisplayName: "Test Region",
			Description: "Test Region Description",
			Types: []corev1.RegionType{
				{
					Name:      "baremetal",
					Locations: []string{"loc-1", "loc-2"},
					Template: &corev1.RegionTemplate{
						Env: []corev1.EnvVar{
							{
								Name:  "ENV_VAR_1",
								Value: "value1",
							},
							{
								Name: "ENV_VAR_2",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &kcorev1.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name: "ENV_VAR_3",
								ValueFrom: &corev1.EnvVarSource{
									ConfigFileKeyRef: &corev1.ConfigFileKeySelector{
										Name: "config-file-name",
									},
								},
							},
						},
						Scheduling: apis.Distributed,
					},
				},
				{
					Name:      "cloud",
					Locations: []string{"loc-3"},
					Template: &corev1.RegionTemplate{
						Env: []corev1.EnvVar{
							{
								Name:  "CLOUD_VAR",
								Value: "cloud-value",
							},
						},
						Scheduling: apis.Packed,
					},
				},
			},
		},
	}

	testRegionModel = regionModel{
		ID:          types.StringValue("test-environment/test-region"),
		Name:        types.StringValue("test-region"),
		Environment: types.StringValue("test-environment"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		DisplayName: types.StringValue("Test Region"),
		Description: types.StringValue("Test Region Description"),
		Types: []regionTypeModel{
			{
				Name: types.StringValue("baremetal"),
				Locations: []types.String{
					types.StringValue("loc-1"),
					types.StringValue("loc-2"),
				},
				Envs: []EnvVarModel{
					{
						Name:  types.StringValue("ENV_VAR_1"),
						Value: types.StringValue("value1"),
					},
					{
						Name:  types.StringValue("ENV_VAR_2"),
						Value: types.StringNull(),
						ValueFrom: &EnvVarSourceModel{
							FieldPath:  types.StringValue("metadata.name"),
							ConfigFile: types.StringNull(),
						},
					},
					{
						Name:  types.StringValue("ENV_VAR_3"),
						Value: types.StringNull(),
						ValueFrom: &EnvVarSourceModel{
							FieldPath:  types.StringNull(),
							ConfigFile: types.StringValue("config-file-name"),
						},
					},
				},
				Scheduling: types.StringValue("Distributed"),
			},
			{
				Name: types.StringValue("cloud"),
				Locations: []types.String{
					types.StringValue("loc-3"),
				},
				Envs: []EnvVarModel{
					{
						Name:  types.StringValue("CLOUD_VAR"),
						Value: types.StringValue("cloud-value"),
					},
				},
				Scheduling: types.StringValue("Packed"),
			},
		},
	}
)
