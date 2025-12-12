package formation

import (
	"testing"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hamba/pkg/v2/ptr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewFormationModel(t *testing.T) {
	model := newFormationModel(testFormationObject)
	assert.Equal(t, testFormationModel, model)
}

func TestFormationModel_ToObject(t *testing.T) {
	obj := testFormationModel.ToObject()
	assert.Equal(t, testFormationObject, obj)
}

var (
	testFormationObject = &formationv1.Formation{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-formation",
			Environment: "test-env",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: formationv1.FormationSpec{
			Description: "Test Formation Description",
			VolumeTemplates: []formationv1.VolumeTemplate{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-volume-template",
					},
					Spec: formationv1.VolumeTemplateSpec{
						VolumeSpec: v1beta1.VolumeSpec{
							VolumeStoreName: "test-volume-store",
							Capacity:        resource.MustParse("5Gi"),
						},
						ReclaimPolicy: formationv1.VolumeTemplateReclaimPolicyRetain,
					},
				},
			},
			Vessels: []formationv1.VesselTemplate{
				{
					Name:        "test-vessel",
					Region:      "test-region",
					Description: "Test Vessel Description",
					Suspend:     ptr.Of(true),
					Override: formationv1.VesselOverride{
						Labels: map[string]string{
							"override-label-key": "override-label-value",
						},
						Containers: []formationv1.VesselContainerOverride{
							{
								Command: []string{"/bin/sh"},
								Args:    []string{"-c", "echo hello"},
								Env: []corev1.EnvVar{
									{
										Name:  "OVERRIDE_ENV",
										Value: "override-value",
									},
								},
							},
						},
					},
				},
			},
			Template: formationv1.GameServerTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"gameserver-label-key": "gameserver-label-value",
					},
					Annotations: map[string]string{
						"gameserver-annotation-key": "gameserver-annotation-value",
					},
				},
				Spec: formationv1.GameServerSpec{
					Containers: []formationv1.Container{
						{
							Name:    "test-container",
							Image:   "test-image",
							Branch:  "test-branch",
							Command: []string{"/app"},
							Args:    []string{"--arg1", "value1"},
							Env: []corev1.EnvVar{
								{
									Name:  "TEST_ENV",
									Value: "test-value",
								},
							},
							Ports: []formationv1.Port{
								{
									Name:               "http",
									Policy:             "Passthrough",
									ContainerPort:      8080,
									Protocol:           kcorev1.ProtocolTCP,
									ProtectionProtocol: ptr.Of("example-protocol"),
								},
							},
							VolumeMounts: []kcorev1.VolumeMount{
								{
									Name:      "test-volume",
									MountPath: "/data",
								},
							},
							Resources: kcorev1.ResourceRequirements{
								Limits: kcorev1.ResourceList{
									kcorev1.ResourceCPU:    resource.MustParse("500m"),
									kcorev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: kcorev1.ResourceList{
									kcorev1.ResourceCPU:    resource.MustParse("250m"),
									kcorev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							Secrets: []formationv1.SecretMount{
								{Name: "secret1", MountPath: "this/is/mount/path"},
								{Name: "secret2", MountPath: "another/path/mount"},
							},
						},
					},
					Health: agonesv1.Health{
						Disabled:            true,
						PeriodSeconds:       10,
						FailureThreshold:    20,
						InitialDelaySeconds: 30,
					},
					TerminationGracePeriodSeconds: ptr.Of[int64](100),
					Volumes: []formationv1.Volume{
						{
							Name: "test-volume",
							Type: formationv1.VolumeTypeEmptyDir,
							EmptyDir: &formationv1.EmptyDirVolumeSource{
								SizeLimit: ptr.Of(resource.MustParse("1Gi")),
							},
						},
					},
					GatewayPolicies: []string{"test-gateway-policy"},
				},
			},
			TerminationGracePeriods: &formationv1.VesselTerminationGracePeriods{
				Maintenance:   ptr.Of[int64](120),
				SpecChange:    ptr.Of[int64](240),
				UserInitiated: ptr.Of[int64](360),
			},
		},
	}

	testFormationModel = formationModel{
		ID:          types.StringValue(cache.NewObjectName("test-env", "test-formation").String()),
		Name:        types.StringValue("test-formation"),
		Environment: types.StringValue("test-env"),
		Description: types.StringValue("Test Formation Description"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		VolumeTemplates: []VolumeTemplateModel{
			{
				Name:            types.StringValue("test-volume-template"),
				ReclaimPolicy:   types.StringValue("Retain"),
				VolumeStoreName: types.StringValue("test-volume-store"),
				Capacity:        types.StringValue("5Gi"),
			},
		},
		Vessels: []VesselTemplateModel{
			{
				Name:        types.StringValue("test-vessel"),
				Region:      types.StringValue("test-region"),
				Description: types.StringValue("Test Vessel Description"),
				Suspend:     types.BoolValue(true),
				Override: &VesselOverrideModel{
					GameServerLabels: map[string]types.String{
						"override-label-key": types.StringValue("override-label-value"),
					},
					Containers: []VesselContainerOverrideModel{
						{
							Command: []types.String{
								types.StringValue("/bin/sh"),
							},
							Args: []types.String{
								types.StringValue("-c"),
								types.StringValue("echo hello"),
							},
							Envs: []core.EnvVarModel{
								{
									Name:  types.StringValue("OVERRIDE_ENV"),
									Value: types.StringValue("override-value"),
								},
							},
						},
					},
				},
			},
		},
		GameServerLabels: map[string]types.String{
			"gameserver-label-key": types.StringValue("gameserver-label-value"),
		},
		GameServerAnnotations: map[string]types.String{
			"gameserver-annotation-key": types.StringValue("gameserver-annotation-value"),
		},
		Containers: []mps.ContainerModel{
			{
				Name: types.StringValue("test-container"),
				ImageRef: mps.ImageRefModel{
					Name:   types.StringValue("test-image"),
					Branch: types.StringValue("test-branch"),
				},
				Command: []types.String{
					types.StringValue("/app"),
				},
				Args: []types.String{
					types.StringValue("--arg1"),
					types.StringValue("value1"),
				},
				Envs: []core.EnvVarModel{
					{
						Name:  types.StringValue("TEST_ENV"),
						Value: types.StringValue("test-value"),
					},
				},
				Ports: []mps.PortModel{
					{
						Name:               types.StringValue("http"),
						Protocol:           types.StringValue("TCP"),
						ContainerPort:      types.Int32Value(8080),
						Policy:             types.StringValue("Passthrough"),
						ProtectionProtocol: types.StringValue("example-protocol"),
					},
				},
				VolumeMounts: []mps.VolumeMountModel{
					{
						Name:      types.StringValue("test-volume"),
						MountPath: types.StringValue("/data"),
					},
				},
				Resources: &mps.ResourcesModel{
					Limits: &mps.ResourceSpecModel{
						CPU:    types.StringValue("500m"),
						Memory: types.StringValue("512Mi"),
					},
					Requests: &mps.ResourceSpecModel{
						CPU:    types.StringValue("250m"),
						Memory: types.StringValue("256Mi"),
					},
				},
				Secrets: []mps.SecretMountModel{
					{Name: types.StringValue("secret1"), MountPath: types.StringValue("this/is/mount/path")},
					{Name: types.StringValue("secret2"), MountPath: types.StringValue("another/path/mount")},
				},
			},
		},
		HealthChecks: &mps.HealthChecksModel{
			Disabled:            types.BoolValue(true),
			PeriodSeconds:       types.Int32Value(10),
			FailureThreshold:    types.Int32Value(20),
			InitialDelaySeconds: types.Int32Value(30),
		},
		TerminationConfig: &terminationConfigModel{
			GracePeriod:   types.Int64Value(100),
			Maintenance:   types.Int64Value(120),
			SpecChange:    types.Int64Value(240),
			UserInitiated: types.Int64Value(360),
		},
		Volumes: []volumeModel{
			{
				Name: types.StringValue("test-volume"),
				EmptyDir: &emptyDirModel{
					SizeLimit: types.StringValue("1Gi"),
				},
			},
		},
		GatewayPolicies: []types.String{
			types.StringValue("test-gateway-policy"),
		},
		ProfilingEnabled:   types.BoolValue(false),
		ImageUpdaterTarget: container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeFormation, "test-formation", "test-env"),
	}
)
