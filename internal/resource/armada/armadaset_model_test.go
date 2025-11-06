package armada

import (
	"testing"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	v1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hamba/pkg/v2/ptr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestNewArmadaSetModel(t *testing.T) {
	model := newArmadaSetModel(testArmadaSetObject)
	assert.Equal(t, testArmadaSetModel, model)
}

func TestArmadaSetModel_ToObject(t *testing.T) {
	obj := testArmadaSetModel.ToObject()
	assert.Equal(t, testArmadaSetObject, obj)
}

var (
	testArmadaSetObject = &armadav1.ArmadaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-armadaset",
			Environment: "test-environment",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: armadav1.ArmadaSetSpec{
			Description: "ArmadaSet Description",
			Armadas: []armadav1.ArmadaTemplate{
				{
					Region: "eu",
					Distribution: []armadav1.ArmadaRegionType{
						{
							Name:        "baremetal",
							MinReplicas: 1,
							MaxReplicas: 5,
							BufferSize:  1,
						},
						{
							Name:        "cloud",
							MinReplicas: 3,
							MaxReplicas: 5,
							BufferSize:  2,
						},
					},
				},
			},
			Override: []armadav1.ArmadaOverride{
				{
					Region: "eu",
					Env: []v1.EnvVar{
						{
							Name:  "ENV_VAR_OVERRIDE",
							Value: "override_value",
						},
					},
					Labels: map[string]string{
						"override-label-key": "override-label-value",
					},
				},
			},
			Autoscaling: armadav1.ArmadaAutoscaling{
				FixedInterval: &armadav1.ArmadaFixInterval{
					Seconds: 60,
				},
			},
			Template: armadav1.FleetTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"gs-label-key":     "gs-label-value",
						"g8c.io/profiling": "true",
					},
					Annotations: map[string]string{
						"gs-annotation-key": "gs-annotation-value",
					},
				},
				Spec: armadav1.FleetSpec{
					GatewayPolicies: []string{
						"policy-1",
						"policy-2",
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &appsv1.RollingUpdateDeployment{
							MaxUnavailable: ptr.Of(intstr.Parse("10%")),
							MaxSurge:       ptr.Of(intstr.Parse("5")),
						},
					},
					Health: agonesv1.Health{
						Disabled:            true,
						PeriodSeconds:       120,
						FailureThreshold:    180,
						InitialDelaySeconds: 10,
					},
					Containers: []armadav1.Container{
						{
							Name:   "container",
							Branch: "branch",
							Image:  "image",
							Ports: []armadav1.Port{
								{
									Name:               "port",
									Policy:             agonesv1.Passthrough,
									ContainerPort:      50000,
									Protocol:           kcorev1.ProtocolUDP,
									ProtectionProtocol: ptr.Of("protection-protocol"),
								},
							},
							Command: []string{"start.sh"},
							Args:    []string{"arg1", "arg2"},
							Env: []v1.EnvVar{
								{
									Name:  "ENV_VAR",
									Value: "value",
								},
								{
									Name: "FIELD_REF_VAR",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &kcorev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "CONFIG_FILE_KEY_REF_VAR",
									ValueFrom: &v1.EnvVarSource{
										ConfigFileKeyRef: &v1.ConfigFileKeySelector{
											Name: "config-file-name",
										},
									},
								},
							},
							ConfigFiles: []armadav1.ConfigFileMount{
								{
									Name:      "config-file",
									MountPath: "/path/to/config",
								},
							},
							Resources: kcorev1.ResourceRequirements{
								Limits: kcorev1.ResourceList{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("256Mi"),
								},
								Requests: kcorev1.ResourceList{
									"cpu":    resource.MustParse("250m"),
									"memory": resource.MustParse("128Mi"),
								},
							},
							VolumeMounts: []kcorev1.VolumeMount{
								{
									Name:      "volume-name",
									MountPath: "/path/to/volume",
								},
							},
						},
					},
					TerminationGracePeriodSeconds: ptr.Of[int64](31),
					Volumes: []armadav1.Volume{
						{
							Name:      "volume-name",
							SizeLimit: ptr.Of(resource.MustParse("10Gi")),
						},
					},
				},
			},
		},
	}

	testArmadaSetModel = armadaSetModel{
		ID:          types.StringValue("test-environment/test-armadaset"),
		Name:        types.StringValue("test-armadaset"),
		Environment: types.StringValue("test-environment"),
		Description: types.StringValue("ArmadaSet Description"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		Autoscaling: &autoscalingModel{
			FixedIntervalSeconds: types.Int32Value(60),
		},
		Regions: []regionModel{
			{
				Name: types.StringValue("eu"),
				Replicas: []replicaModel{
					{
						RegionType:  types.StringValue("baremetal"),
						MinReplicas: types.Int32Value(1),
						MaxReplicas: types.Int32Value(5),
						BufferSize:  types.Int32Value(1),
					},
					{
						RegionType:  types.StringValue("cloud"),
						MinReplicas: types.Int32Value(3),
						MaxReplicas: types.Int32Value(5),
						BufferSize:  types.Int32Value(2),
					},
				},
				Envs: []core.EnvVarModel{
					{
						Name:  types.StringValue("ENV_VAR_OVERRIDE"),
						Value: types.StringValue("override_value"),
					},
				},
				GameServerLabels: map[string]types.String{
					"override-label-key": types.StringValue("override-label-value"),
				},
			},
		},
		GameServerLabels: map[string]types.String{
			"gs-label-key": types.StringValue("gs-label-value"),
		},
		GameServerAnnotations: map[string]types.String{
			"gs-annotation-key": types.StringValue("gs-annotation-value"),
		},
		Containers: []mps.ContainerModel{
			{
				Name: types.StringValue("container"),
				ImageRef: mps.ImageRefModel{
					Name:   types.StringValue("image"),
					Branch: types.StringValue("branch"),
				},
				Ports: []mps.PortModel{
					{
						Name:               types.StringValue("port"),
						Policy:             types.StringValue(string(agonesv1.Passthrough)),
						ContainerPort:      types.Int32Value(50000),
						Protocol:           types.StringValue(string(kcorev1.ProtocolUDP)),
						ProtectionProtocol: types.StringValue("protection-protocol"),
					},
				},
				Command: []types.String{
					types.StringValue("start.sh"),
				},
				Args: []types.String{
					types.StringValue("arg1"),
					types.StringValue("arg2"),
				},
				Envs: []core.EnvVarModel{
					{
						Name:  types.StringValue("ENV_VAR"),
						Value: types.StringValue("value"),
					},
					{
						Name: types.StringValue("FIELD_REF_VAR"),
						ValueFrom: &core.EnvVarSourceModel{
							FieldPath: types.StringValue("metadata.name"),
						},
					},
					{
						Name: types.StringValue("CONFIG_FILE_KEY_REF_VAR"),
						ValueFrom: &core.EnvVarSourceModel{
							ConfigFile: types.StringValue("config-file-name"),
						},
					},
				},
				ConfigFiles: []mps.ConfigFileModel{
					{
						Name:      types.StringValue("config-file"),
						MountPath: types.StringValue("/path/to/config"),
					},
				},
				Resources: &mps.ResourcesModel{
					Limits: &mps.ResourceSpecModel{
						CPU:    types.StringValue("500m"),
						Memory: types.StringValue("256Mi"),
					},
					Requests: &mps.ResourceSpecModel{
						CPU:    types.StringValue("250m"),
						Memory: types.StringValue("128Mi"),
					},
				},
				VolumeMounts: []mps.VolumeMountModel{
					{
						Name:      types.StringValue("volume-name"),
						MountPath: types.StringValue("/path/to/volume"),
					},
				},
			},
		},
		HealthChecks: &mps.HealthChecksModel{
			Disabled:            types.BoolValue(true),
			PeriodSeconds:       types.Int32Value(120),
			FailureThreshold:    types.Int32Value(180),
			InitialDelaySeconds: types.Int32Value(10),
		},
		TerminationConfig: &terminationConfigModel{
			GracePeriodSeconds: types.Int64Value(31),
		},
		Strategy: &strategyModel{
			RollingUpdate: &rollingUpdateModel{
				MaxUnavailable: types.StringValue("10%"),
				MaxSurge:       types.StringValue("5"),
			},
		},
		Volumes: []volumeModel{
			{
				Name: types.StringValue("volume-name"),
				EmptyDir: &emptyDirModel{
					SizeLimit: types.StringValue("10Gi"),
				},
			},
		},
		GatewayPolicies: []types.String{
			types.StringValue("policy-1"),
			types.StringValue("policy-2"),
		},
		ProfilingEnabled:   types.BoolValue(true),
		ImageUpdaterTarget: container.NewImageUpdaterTargetModel("armadaset", "test-armadaset", "test-environment"),
	}
)
