package armada

import (
	"strconv"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/hashicorp/terraform-plugin-framework/types"
	appsv1 "k8s.io/api/apps/v1"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const profilingAnnotation = "g8c.io/profiling"

type armadaModel struct {
	ID                    types.String            `tfsdk:"id"`
	Name                  types.String            `tfsdk:"name"`
	Environment           types.String            `tfsdk:"environment"`
	Description           types.String            `tfsdk:"description"`
	Labels                map[string]types.String `tfsdk:"labels"`
	Annotations           map[string]types.String `tfsdk:"annotations"`
	Autoscaling           *Autoscaling            `tfsdk:"autoscaling"`
	Region                types.String            `tfsdk:"region"`
	Replicas              []Replica               `tfsdk:"replicas"`
	GameServerLabels      map[string]types.String `tfsdk:"gameserver_labels"`
	GameServerAnnotations map[string]types.String `tfsdk:"gameserver_annotations"`
	Containers            []Container             `tfsdk:"containers"`
	HealthChecks          HealthChecks            `tfsdk:"health_checks"`
	TerminationConfig     TerminationConfig       `tfsdk:"termination_configuration"`
	Strategy              Strategy                `tfsdk:"strategy"`
	Volumes               []Volume                `tfsdk:"volumes"`
	GatewayPolicies       []types.String          `tfsdk:"gateway_policies"`
	ProfilingEnabled      types.Bool              `tfsdk:"profiling_enabled"`
	ImageUpdaterTarget    ImageUpdaterTarget      `tfsdk:"image_updater_target"`
}

func newArmadaModel(obj *armadav1.Armada) armadaModel {
	return armadaModel{
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Description:           types.StringValue(obj.Spec.Description),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           conv.ForEachMapItem(obj.Annotations, types.StringValue),
		Autoscaling:           newAutoscalingModel(obj.Spec.Autoscaling),
		Region:                types.StringValue(obj.Spec.Region),
		Replicas:              conv.ForEachSliceItem(obj.Spec.Distribution, newReplicas),
		GameServerLabels:      conv.ForEachMapItem(obj.Spec.Template.Labels, types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, newContainer),
		HealthChecks: HealthChecks{
			Disabled:            types.BoolValue(obj.Spec.Template.Spec.Health.Disabled),
			InitialDelaySeconds: types.Int32Value(obj.Spec.Template.Spec.Health.InitialDelaySeconds),
			PeriodSeconds:       types.Int32Value(obj.Spec.Template.Spec.Health.PeriodSeconds),
			FailureThreshold:    types.Int32Value(obj.Spec.Template.Spec.Health.FailureThreshold),
		},
		TerminationConfig: TerminationConfig{
			GracePeriodSeconds: conv.OptionalFunc(*obj.Spec.Template.Spec.TerminationGracePeriodSeconds, types.Int64Value, types.Int64Null),
		},
		Strategy:         newStrategy(obj.Spec.Template.Spec.Strategy),
		Volumes:          conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolume),
		GatewayPolicies:  conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled: newProfilingEnabled(obj.Annotations),
	}
}

func (m armadaModel) ToObject() *armadav1.Armada {
	return &armadav1.Armada{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(v types.String) string { return v.ValueString() }),
		},
		Spec: armadav1.ArmadaSpec{
			Description: m.Description.ValueString(),
			Region:      m.Region.ValueString(),
			Distribution: conv.ForEachSliceItem(m.Replicas, func(r Replica) armadav1.ArmadaRegionType {
				return armadav1.ArmadaRegionType{
					Name:        r.RegionType.ValueString(),
					MinReplicas: r.MinReplicas.ValueInt32(),
					MaxReplicas: r.MaxReplicas.ValueInt32(),
					BufferSize:  r.BufferSize.ValueInt32(),
				}
			}),
			Autoscaling: armadav1.ArmadaAutoscaling{
				FixedInterval: toFixedInterval(m.Autoscaling.FixedIntervalSeconds),
			},
			Template: armadav1.FleetTemplateSpec{
				Spec: armadav1.FleetSpec{
					GatewayPolicies:               conv.ForEachSliceItem(m.GatewayPolicies, func(v types.String) string { return v.ValueString() }),
					Strategy:                      toDeploymentStrategy(m.Strategy),
					Health:                        toHealthChecks(m.HealthChecks),
					Containers:                    conv.ForEachSliceItem(m.Containers, toContainer),
					TerminationGracePeriodSeconds: toTerminationGracePeriodSeconds(m.TerminationConfig.GracePeriodSeconds),
					Volumes:                       conv.ForEachSliceItem(m.Volumes, toVolume),
				},
			},
		},
	}
}

func toVolume(vol Volume) armadav1.Volume {
	res := armadav1.Volume{
		Name: vol.Name.ValueString(),
	}
	if vol.EmptyDir != nil {
		q := toQuantity(vol.EmptyDir.SizeLimit)
		res.SizeLimit = &q
	}
	return res
}

func toContainer(ctr Container) armadav1.Container {
	return armadav1.Container{
		Name:         ctr.Name.ValueString(),
		Image:        ctr.Image.Name.ValueString(),
		Branch:       ctr.Image.Branch.ValueString(),
		Command:      conv.ForEachSliceItem(ctr.Command, func(v types.String) string { return v.ValueString() }),
		Args:         conv.ForEachSliceItem(ctr.Args, func(v types.String) string { return v.ValueString() }),
		Resources:    toResourceRequirements(ctr.Resources),
		Env:          conv.ForEachSliceItem(ctr.Envs, func(item core.EnvVarModel) corev1.EnvVar { return item.ToObject() }),
		Ports:        conv.ForEachSliceItem(ctr.Ports, toPort),
		VolumeMounts: conv.ForEachSliceItem(ctr.VolumeMounts, toVolumeMount),
		ConfigFiles:  conv.ForEachSliceItem(ctr.ConfigFiles, toConfigFile),
	}
}

func toConfigFile(file ConfigFile) armadav1.ConfigFileMount {
	return armadav1.ConfigFileMount{
		Name:      file.Name.ValueString(),
		MountPath: file.MountPath.ValueString(),
	}
}

func toVolumeMount(mount VolumeMount) kcorev1.VolumeMount {
	return kcorev1.VolumeMount{
		Name:        mount.Name.ValueString(),
		MountPath:   mount.MountPath.ValueString(),
		SubPath:     mount.SubPath.ValueString(),
		SubPathExpr: mount.SubPathExpr.ValueString(),
	}
}

func toPort(p Port) armadav1.Port {
	return armadav1.Port{
		Name:               p.Name.ValueString(),
		Policy:             agonesv1.PortPolicy(p.Policy.ValueString()),
		ContainerPort:      uint16(p.ContainerPort.ValueInt32()),
		Protocol:           kcorev1.Protocol(p.Protocol.ValueString()),
		ProtectionProtocol: p.ProtectionProtocol.ValueStringPointer(),
	}
}

func toResourceRequirements(res *Resources) kcorev1.ResourceRequirements {
	if res == nil {
		return kcorev1.ResourceRequirements{}
	}
	return kcorev1.ResourceRequirements{
		Limits:   toResourceList(res.Limits),
		Requests: toResourceList(res.Requests),
	}
}

func toResourceList(spec ResourceSpec) kcorev1.ResourceList {
	res := kcorev1.ResourceList{}
	if q := toQuantity(spec.CPU); !q.IsZero() {
		res[kcorev1.ResourceCPU] = q
	}
	if q := toQuantity(spec.Memory); !q.IsZero() {
		res[kcorev1.ResourceMemory] = q
	}
	return res
}

func toQuantity(val types.String) resource.Quantity {
	if val.IsNull() || val.IsUnknown() {
		return resource.Quantity{}
	}
	return resource.MustParse(val.ValueString()) // Pre-validation required.
}

func toTerminationGracePeriodSeconds(seconds types.Int64) *int64 {
	if seconds.IsNull() || seconds.IsUnknown() {
		return nil
	}
	return seconds.ValueInt64Pointer()
}

func toHealthChecks(hc HealthChecks) agonesv1.Health {
	return agonesv1.Health{
		Disabled:            hc.Disabled.ValueBool(),
		InitialDelaySeconds: hc.InitialDelaySeconds.ValueInt32(),
		PeriodSeconds:       hc.PeriodSeconds.ValueInt32(),
		FailureThreshold:    hc.FailureThreshold.ValueInt32(),
	}
}

func toDeploymentStrategy(strat Strategy) appsv1.DeploymentStrategy {
	if strat.RollingUpdate != nil {
		return appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxSurge:       toIntOrString(strat.RollingUpdate.MaxSurge),
				MaxUnavailable: toIntOrString(strat.RollingUpdate.MaxUnavailable),
			},
		}
	}
	return appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	}
}

func toIntOrString(val types.String) *intstr.IntOrString {
	if val.IsNull() || val.IsUnknown() {
		return nil
	}
	is := intstr.Parse(val.ValueString())
	return &is
}

func toFixedInterval(val types.Int32) *armadav1.ArmadaFixInterval {
	if val.IsNull() || val.IsUnknown() {
		return &armadav1.ArmadaFixInterval{}
	}
	return &armadav1.ArmadaFixInterval{
		Seconds: val.ValueInt32(),
	}
}

func newProfilingEnabled(annots map[string]string) types.Bool {
	if annots == nil {
		return types.BoolNull()
	}

	val, known := annots[profilingAnnotation]
	if !known {
		return types.BoolNull()
	}
	return types.BoolValue(val == "true")
}

type Autoscaling struct {
	FixedIntervalSeconds types.Int32 `tfsdk:"fixed_interval_seconds"`
}

func newAutoscalingModel(obj armadav1.ArmadaAutoscaling) *Autoscaling {
	if obj.FixedInterval == nil {
		return nil
	}
	return &Autoscaling{
		FixedIntervalSeconds: types.Int32Value(obj.FixedInterval.Seconds),
	}
}

type Replica struct {
	RegionType  types.String `tfsdk:"region_type"`
	MinReplicas types.Int32  `tfsdk:"min_replicas"`
	MaxReplicas types.Int32  `tfsdk:"max_replicas"`
	BufferSize  types.Int32  `tfsdk:"buffer_size"`
}

func newReplicas(obj armadav1.ArmadaRegionType) Replica {
	return Replica{
		RegionType:  types.StringValue(obj.Name),
		MinReplicas: types.Int32Value(obj.MinReplicas),
		MaxReplicas: types.Int32Value(obj.MaxReplicas),
		BufferSize:  types.Int32Value(obj.BufferSize),
	}
}

type Container struct {
	Name         types.String       `tfsdk:"name"`
	Image        Image              `tfsdk:"image"`
	Command      []types.String     `tfsdk:"command"`
	Args         []types.String     `tfsdk:"args"`
	Resources    *Resources         `tfsdk:"resources"`
	Envs         []core.EnvVarModel `tfsdk:"envs"`
	Ports        []Port             `tfsdk:"ports"`
	VolumeMounts []VolumeMount      `tfsdk:"volume_mounts"`
	ConfigFiles  []ConfigFile       `tfsdk:"config_files"`
}

func newContainer(obj armadav1.Container) Container {
	return Container{
		Name: types.StringValue(obj.Name),
		Image: Image{
			Name:   types.StringValue(obj.Image),
			Branch: types.StringValue(obj.Branch),
		},
		Command:      conv.ForEachSliceItem(obj.Command, types.StringValue),
		Args:         conv.ForEachSliceItem(obj.Args, types.StringValue),
		Resources:    newResources(obj.Resources),
		Envs:         conv.ForEachSliceItem(obj.Env, core.NewEnvVarModel),
		Ports:        conv.ForEachSliceItem(obj.Ports, newPort),
		VolumeMounts: conv.ForEachSliceItem(obj.VolumeMounts, newVolumeMounts),
		ConfigFiles:  conv.ForEachSliceItem(obj.ConfigFiles, newConfigFiles),
	}
}

type Image struct {
	Name   types.String `tfsdk:"name"`
	Branch types.String `tfsdk:"branch"`
}

type Resources struct {
	Limits   ResourceSpec `tfsdk:"limits"`
	Requests ResourceSpec `tfsdk:"requests"`
}

func newResources(obj kcorev1.ResourceRequirements) *Resources {
	res := &Resources{}
	if obj.Limits != nil {
		res.Limits = newResourceSpec(obj.Limits)
	}
	if obj.Requests != nil {
		res.Requests = newResourceSpec(obj.Requests)
	}
	return res
}

func newResourceSpec(obj kcorev1.ResourceList) ResourceSpec {
	res := ResourceSpec{}
	if obj.Cpu() != nil {
		res.CPU = types.StringValue(obj.Cpu().String())
	}
	if obj.Memory() != nil {
		res.Memory = types.StringValue(obj.Memory().String())
	}
	return res
}

type ResourceSpec struct {
	CPU    types.String `tfsdk:"cpu"`
	Memory types.String `tfsdk:"memory"`
}

type Port struct {
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	ContainerPort      types.Int32  `tfsdk:"container_port"`
	Policy             types.String `tfsdk:"policy"`
	ProtectionProtocol types.String `tfsdk:"protection_protocol"`
}

func newPort(obj armadav1.Port) Port {
	return Port{
		Name:               types.StringValue(obj.Name),
		Protocol:           types.StringValue(string(obj.Protocol)),
		ContainerPort:      types.Int32Value(int32(obj.ContainerPort)),
		Policy:             types.StringValue(string(obj.Policy)),
		ProtectionProtocol: conv.OptionalFunc(*obj.ProtectionProtocol, types.StringValue, types.StringNull),
	}
}

type VolumeMount struct {
	Name        types.String `tfsdk:"name"`
	MountPath   types.String `tfsdk:"mount_path"`
	SubPath     types.String `tfsdk:"sub_path"`
	SubPathExpr types.String `tfsdk:"sub_path_expr"`
}

func newVolumeMounts(obj kcorev1.VolumeMount) VolumeMount {
	return VolumeMount{
		Name:        types.StringValue(obj.Name),
		MountPath:   types.StringValue(obj.MountPath),
		SubPath:     conv.OptionalFunc(obj.SubPath, types.StringValue, types.StringNull),
		SubPathExpr: conv.OptionalFunc(obj.SubPathExpr, types.StringValue, types.StringNull),
	}
}

type ConfigFile struct {
	Name      types.String `tfsdk:"name"`
	MountPath types.String `tfsdk:"mount_path"`
}

func newConfigFiles(obj armadav1.ConfigFileMount) ConfigFile {
	return ConfigFile{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

type HealthChecks struct {
	Disabled            types.Bool  `tfsdk:"disabled"`
	InitialDelaySeconds types.Int32 `tfsdk:"initial_delay_seconds"`
	PeriodSeconds       types.Int32 `tfsdk:"period_seconds"`
	FailureThreshold    types.Int32 `tfsdk:"failure_threshold"`
}

type TerminationConfig struct {
	GracePeriodSeconds   types.Int64 `tfsdk:"grace_period_seconds"`
	MaintenanceSeconds   types.Int32 `tfsdk:"maintenance_seconds"`
	SpecChangeSeconds    types.Int32 `tfsdk:"spec_change_seconds"`
	UserInitiatedSeconds types.Int32 `tfsdk:"user_initiated_seconds"`
}

type Strategy struct {
	RollingUpdate *RollingUpdate `tfsdk:"rolling_update"`
	Recreate      types.Object   `tfsdk:"recreate"`
}

func newStrategy(obj appsv1.DeploymentStrategy) Strategy {
	switch obj.Type {
	case appsv1.RollingUpdateDeploymentStrategyType:
		if obj.RollingUpdate == nil {
			return Strategy{
				RollingUpdate: &RollingUpdate{},
			}
		}
		return Strategy{
			RollingUpdate: &RollingUpdate{
				MaxSurge:       newIntOrString(obj.RollingUpdate.MaxSurge),
				MaxUnavailable: newIntOrString(obj.RollingUpdate.MaxUnavailable),
			},
		}
	default:
		return Strategy{
			Recreate: types.Object{},
		}
	}
}

func newIntOrString(val *intstr.IntOrString) types.String {
	if val == nil {
		return types.StringNull()
	}
	if val.Type == intstr.Int {
		return types.StringValue(strconv.Itoa(int(val.IntVal)))
	}
	return types.StringValue(val.StrVal)
}

type RollingUpdate struct {
	MaxSurge       types.String `tfsdk:"max_surge"`
	MaxUnavailable types.String `tfsdk:"max_unavailable"`
}

type Volume struct {
	Name     types.String `tfsdk:"name"`
	EmptyDir *EmptyDir    `tfsdk:"empty_dir"`
}

func newVolume(obj armadav1.Volume) Volume {
	vol := Volume{
		Name: types.StringValue(obj.Name),
	}
	if vol.EmptyDir != nil {
		vol.EmptyDir = &EmptyDir{
			SizeLimit: conv.OptionalFunc(obj.SizeLimit.String(), types.StringValue, types.StringNull),
		}
	}
	return vol
}

type EmptyDir struct {
	SizeLimit types.String `tfsdk:"size_limit"`
}

type ImageUpdaterTarget struct {
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Environment types.String `tfsdk:"environment"`
}
