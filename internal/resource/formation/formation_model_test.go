package formation

import (
	"testing"

	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	v1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
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
	testGracePeriod       = int64(30)
	testMaintenance       = int64(60)
	testSpecChange        = int64(90)
	testUserInitiated     = int64(120)
	testSuspend           = false
	testQuantity5Gi       = resource.MustParse("5Gi")
	testSizeLimitQuantity = resource.MustParse("1Gi")

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
							Capacity:        testQuantity5Gi,
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
					Suspend:     &testSuspend,
					Override: formationv1.VesselOverride{
						Labels: map[string]string{
							"override-label-key": "override-label-value",
						},
						Containers: []formationv1.VesselContainerOverride{
							{
								Command: []string{"/bin/sh"},
								Args:    []string{"-c", "echo hello"},
								Env: []v1.EnvVar{
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
					Containers: []v1.Container{
						{
							Name: "test-container",
							ImageRef: v1.ImageRef{
								Name:   "test-image",
								Branch: "test-branch",
							},
							Command: []string{"/app"},
							Args:    []string{"--arg1", "value1"},
							Env: []v1.EnvVar{
								{
									Name:  "TEST_ENV",
									Value: "test-value",
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "test-volume",
									MountPath: "/data",
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("250m"),
									v1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
					Health: &formationv1.HealthChecks{
						Startup: &formationv1.HealthCheck{
							Enabled: true,
							Probe: v1.Probe{
								HTTPGet: &v1.HTTPGetAction{
									Port: v1.IntOrString{IntVal: 8080},
									Path: "/startup",
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
						},
					},
					TerminationGracePeriodSeconds: &testGracePeriod,
					Volumes: []formationv1.Volume{
						{
							Name: "test-volume",
							Type: formationv1.VolumeTypeEmptyDir,
							EmptyDir: &formationv1.EmptyDirVolumeSource{
								SizeLimit: &testSizeLimitQuantity,
							},
						},
					},
					GatewayPolicies: []string{"test-gateway-policy"},
				},
			},
			TerminationGracePeriods: &formationv1.VesselTerminationGracePeriods{
				Maintenance:   &testMaintenance,
				SpecChange:    &testSpecChange,
				UserInitiated: &testUserInitiated,
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
				Suspend:     types.BoolValue(false),
				Override: &VesselOverrideModel{
					Labels: map[string]types.String{
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
				ImageRef: &container.ImageRefModel{
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
				Ports: []mps.ContainerPortModel{
					{
						Name:          types.StringValue("http"),
						Protocol:      types.StringValue("TCP"),
						ContainerPort: types.Int64Value(8080),
					},
				},
				VolumeMounts: []mps.VolumeMountModel{
					{
						Name:      types.StringValue("test-volume"),
						MountPath: types.StringValue("/data"),
					},
				},
				Resources: &mps.ResourceRequirementsModel{
					Limits: &mps.ResourceListModel{
						CPU:    types.StringValue("500m"),
						Memory: types.StringValue("512Mi"),
					},
					Requests: &mps.ResourceListModel{
						CPU:    types.StringValue("250m"),
						Memory: types.StringValue("256Mi"),
					},
				},
			},
		},
		HealthChecks: &mps.HealthChecksModel{
			Startup: &mps.HealthCheckModel{
				Enabled: types.BoolValue(true),
				HTTPGet: &mps.HTTPGetActionModel{
					Port: types.Int64Value(8080),
					Path: types.StringValue("/startup"),
				},
				InitialDelaySeconds: types.Int64Value(5),
				PeriodSeconds:       types.Int64Value(10),
			},
		},
		TerminationConfig: &terminationConfigModel{
			GracePeriod:   types.Int64Value(30),
			Maintenance:   types.Int64Value(60),
			SpecChange:    types.Int64Value(90),
			UserInitiated: types.Int64Value(120),
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
