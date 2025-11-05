package formation

import (
	"testing"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewVesselModel(t *testing.T) {
	model := newVesselModel(testVesselObject)
	assert.Equal(t, testVesselModel, model)
}

func TestVesselModel_ToObject(t *testing.T) {
	obj := testVesselModel.ToObject()
	assert.Equal(t, testVesselObject, obj)
}

var (
	testVesselObject = &formationv1.Vessel{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-vessel",
			Environment: "test-env",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: formationv1.VesselSpec{
			Description: "Test Vessel Description",
			Region:      "us-west",
			Suspend:     boolPtr(true),
			Template: formationv1.GameServerTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"gameserver-label-key": "gameserver-label-value",
						profilingKey:           "true",
					},
					Annotations: map[string]string{
						"gameserver-annotation-key": "gameserver-annotation-value",
					},
				},
				Spec: formationv1.GameServerSpec{
					GatewayPolicies: []string{
						"gateway-policy-1",
						"gateway-policy-2",
					},
					Health: agonesv1.Health{
						Disabled:            false,
						InitialDelaySeconds: int32(10),
						PeriodSeconds:       int32(5),
						FailureThreshold:    int32(3),
					},
					Containers: []formationv1.Container{
						{
							Name:   "game-server",
							Image:  "game-image",
							Branch: "main",
							Command: []string{
								"/bin/sh",
							},
							Args: []string{
								"-c",
								"echo hello",
							},
							Resources: kcorev1.ResourceRequirements{
								Limits: kcorev1.ResourceList{
									kcorev1.ResourceCPU:    resource.MustParse("1000m"),
									kcorev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: kcorev1.ResourceList{
									kcorev1.ResourceCPU:    resource.MustParse("500m"),
									kcorev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENV_VAR_1",
									Value: "value1",
								},
							},
							Ports: []formationv1.ContainerPort{
								{
									Name:          "game-port",
									ContainerPort: int32(7777),
									Protocol:      kcorev1.ProtocolUDP,
									HostPort:      int32(0),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
							ConfigFiles: []formationv1.ConfigFileMount{
								{
									Name:        "config-file-1",
									MountPath:   "/config/file1.conf",
									Content:     "config content",
									Permissions: "0644",
								},
							},
						},
					},
					TerminationGracePeriodSeconds: int64Ptr(30),
					Volumes: []formationv1.Volume{
						{
							Name: "data-volume",
							Type: formationv1.VolumeTypeEmptyDir,
							EmptyDir: &formationv1.EmptyDirVolumeSource{
								SizeLimit: resource.MustParse("1Gi"),
							},
						},
						{
							Name: "persistent-volume",
							Type: formationv1.VolumeTypePersistent,
							Persistent: &formationv1.PersistentVolumeSource{
								VolumeName: "my-persistent-volume",
							},
						},
					},
				},
			},
			TerminationGracePeriods: &formationv1.VesselTerminationGracePeriods{
				Maintenance:   int64Ptr(60),
				SpecChange:    int64Ptr(45),
				UserInitiated: int64Ptr(90),
			},
		},
	}

	testVesselModel = vesselModel{
		ID:          types.StringValue("test-env/test-vessel"),
		Name:        types.StringValue("test-vessel"),
		Environment: types.StringValue("test-env"),
		Region:      types.StringValue("us-west"),
		Description: types.StringValue("Test Vessel Description"),
		Suspend:     types.BoolPointerValue(boolPtr(true)),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		GameServerLabels: map[string]types.String{
			"gameserver-label-key": types.StringValue("gameserver-label-value"),
		},
		GameServerAnnotations: map[string]types.String{
			"gameserver-annotation-key": types.StringValue("gameserver-annotation-value"),
		},
		GatewayPolicies: []types.String{
			types.StringValue("gateway-policy-1"),
			types.StringValue("gateway-policy-2"),
		},
		ProfilingEnabled: types.BoolValue(true),
		HealthChecks: &mps.HealthChecksModel{
			Disabled:            types.BoolValue(false),
			InitialDelaySeconds: types.Int32Value(10),
			PeriodSeconds:       types.Int32Value(5),
			FailureThreshold:    types.Int32Value(3),
		},
		Containers: []mps.ContainerModel{
			{
				Name: types.StringValue("game-server"),
				ImageRef: mps.ImageRefModel{
					Name:   types.StringValue("game-image"),
					Branch: types.StringValue("main"),
				},
				Command: []types.String{
					types.StringValue("/bin/sh"),
				},
				Args: []types.String{
					types.StringValue("-c"),
					types.StringValue("echo hello"),
				},
				Resources: &mps.ResourcesModel{
					Limits: &mps.ResourceModel{
						CPU:    types.StringValue("1000m"),
						Memory: types.StringValue("512Mi"),
					},
					Requests: &mps.ResourceModel{
						CPU:    types.StringValue("500m"),
						Memory: types.StringValue("256Mi"),
					},
				},
				Envs: []core.EnvVarModel{
					{
						Name:  types.StringValue("ENV_VAR_1"),
						Value: types.StringValue("value1"),
					},
				},
				Ports: []mps.PortModel{
					{
						Name:          types.StringValue("game-port"),
						ContainerPort: types.Int32Value(7777),
						Protocol:      types.StringValue("UDP"),
						HostPort:      types.Int32Value(0),
					},
				},
				VolumeMounts: []mps.VolumeMountModel{
					{
						Name:      types.StringValue("data-volume"),
						MountPath: types.StringValue("/data"),
					},
				},
				ConfigFiles: []mps.ConfigFileModel{
					{
						Name:        types.StringValue("config-file-1"),
						MountPath:   types.StringValue("/config/file1.conf"),
						Content:     types.StringValue("config content"),
						Permissions: types.StringValue("0644"),
					},
				},
			},
		},
		TerminationConfig: &terminationConfigModel{
			GracePeriod:   types.Int64PointerValue(int64Ptr(30)),
			Maintenance:   types.Int64PointerValue(int64Ptr(60)),
			SpecChange:    types.Int64PointerValue(int64Ptr(45)),
			UserInitiated: types.Int64PointerValue(int64Ptr(90)),
		},
		Volumes: []volumeModel{
			{
				Name: types.StringValue("data-volume"),
				EmptyDir: &emptyDirModel{
					SizeLimit: types.StringValue("1Gi"),
				},
			},
			{
				Name: types.StringValue("persistent-volume"),
				Persistent: &persistentModel{
					VolumeName: types.StringValue("my-persistent-volume"),
				},
			},
		},
		ImageUpdaterTarget: container.NewImageUpdaterTargetModel(
			container.ImageUpdaterTargetTypeVessel,
			"test-vessel",
			"test-env",
		),
	}
)

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
