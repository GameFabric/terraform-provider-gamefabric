package armada

import (
	"maps"
	"strconv"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/hashicorp/terraform-plugin-framework/types"
	appsv1 "k8s.io/api/apps/v1"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const profilingAnnotation = "g8c.io/profiling"

type armadaModel struct {
	ID                    types.String                       `tfsdk:"id"`
	Name                  types.String                       `tfsdk:"name"`
	Environment           types.String                       `tfsdk:"environment"`
	Description           types.String                       `tfsdk:"description"`
	Labels                map[string]types.String            `tfsdk:"labels"`
	Annotations           map[string]types.String            `tfsdk:"annotations"`
	Autoscaling           *autoscalingModel                  `tfsdk:"autoscaling"`
	Region                types.String                       `tfsdk:"region"`
	Replicas              []replicaModel                     `tfsdk:"replicas"`
	GameServerLabels      map[string]types.String            `tfsdk:"gameserver_labels"`
	GameServerAnnotations map[string]types.String            `tfsdk:"gameserver_annotations"`
	Containers            []containerModel                   `tfsdk:"containers"`
	HealthChecks          *healthChecksModel                 `tfsdk:"health_checks"`
	TerminationConfig     *terminationConfigModel            `tfsdk:"termination_configuration"`
	Strategy              *strategyModel                     `tfsdk:"strategy"`
	Volumes               []volumeModel                      `tfsdk:"volumes"`
	GatewayPolicies       []types.String                     `tfsdk:"gateway_policies"`
	ProfilingEnabled      types.Bool                         `tfsdk:"profiling_enabled"`
	ImageUpdaterTarget    *container.ImageUpdaterTargetModel `tfsdk:"image_updater_target"`
}

func newArmadaModel(obj *armadav1.Armada) armadaModel {
	annots := conv.ForEachMapItem(obj.Annotations, types.StringValue)
	delete(annots, profilingAnnotation)

	return armadaModel{
		ID:                    types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Description:           conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           annots,
		Autoscaling:           newAutoscalingModel(obj.Spec.Autoscaling),
		Region:                types.StringValue(obj.Spec.Region),
		Replicas:              conv.ForEachSliceItem(obj.Spec.Distribution, newReplicas),
		GameServerLabels:      conv.ForEachMapItem(obj.Spec.Template.Labels, types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, newContainer),
		HealthChecks:          newHealthChecks(obj.Spec.Template.Spec.Health),
		TerminationConfig:     newTerminationConfig(obj.Spec.Template.Spec.TerminationGracePeriodSeconds),
		Strategy:              newStrategyModel(obj.Spec.Template.Spec.Strategy),
		Volumes:               conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolumeModel),
		GatewayPolicies:       conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled:      newProfilingEnabled(obj.Annotations),
		ImageUpdaterTarget:    container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeArmada, obj.Name, obj.Environment),
	}
}

func newTerminationConfig(seconds *int64) *terminationConfigModel {
	if seconds == nil {
		return nil
	}
	return &terminationConfigModel{
		GracePeriodSeconds: conv.OptionalFunc(*seconds, types.Int64Value, types.Int64Null),
	}
}

func newHealthChecks(obj agonesv1.Health) *healthChecksModel {
	if obj == (agonesv1.Health{}) {
		return nil
	}

	return &healthChecksModel{
		Disabled:            types.BoolValue(obj.Disabled),
		InitialDelaySeconds: types.Int32Value(obj.InitialDelaySeconds),
		PeriodSeconds:       types.Int32Value(obj.PeriodSeconds),
		FailureThreshold:    types.Int32Value(obj.FailureThreshold),
	}
}

func (m armadaModel) ToObject() *armadav1.Armada {
	return &armadav1.Armada{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(
				toAnnotations(m.Annotations, m.ProfilingEnabled),
				func(v types.String) string { return v.ValueString() },
			),
		},
		Spec: armadav1.ArmadaSpec{
			Description: m.Description.ValueString(),
			Region:      m.Region.ValueString(),
			Distribution: conv.ForEachSliceItem(m.Replicas, func(r replicaModel) armadav1.ArmadaRegionType {
				return armadav1.ArmadaRegionType{
					Name:        r.RegionType.ValueString(),
					MinReplicas: r.MinReplicas.ValueInt32(),
					MaxReplicas: r.MaxReplicas.ValueInt32(),
					BufferSize:  r.BufferSize.ValueInt32(),
				}
			}),
			Autoscaling: armadav1.ArmadaAutoscaling{
				FixedInterval: toFixedInterval(m.Autoscaling),
			},
			Template: armadav1.FleetTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      conv.ForEachMapItem(m.GameServerLabels, func(v types.String) string { return v.ValueString() }),
					Annotations: conv.ForEachMapItem(m.GameServerAnnotations, func(v types.String) string { return v.ValueString() }),
				},
				Spec: armadav1.FleetSpec{
					GatewayPolicies:               conv.ForEachSliceItem(m.GatewayPolicies, func(v types.String) string { return v.ValueString() }),
					Strategy:                      toDeploymentStrategy(m.Strategy),
					Health:                        toHealthChecks(m.HealthChecks),
					Containers:                    conv.ForEachSliceItem(m.Containers, toContainer),
					TerminationGracePeriodSeconds: toTerminationGracePeriodSeconds(m.TerminationConfig),
					Volumes:                       conv.ForEachSliceItem(m.Volumes, toVolume),
				},
			},
		},
	}
}

func toAnnotations(annots map[string]types.String, profilingEnabled types.Bool) map[string]types.String {
	if conv.IsKnown(profilingEnabled) {
		res := maps.Clone(annots)
		res[profilingAnnotation] = types.StringValue(strconv.FormatBool(profilingEnabled.ValueBool()))
		return res
	}
	return annots
}

func toVolume(vol volumeModel) armadav1.Volume {
	res := armadav1.Volume{
		Name: vol.Name.ValueString(),
	}
	if vol.EmptyDir != nil {
		q := toQuantity(vol.EmptyDir.SizeLimit)
		res.SizeLimit = &q
	}
	return res
}

func toContainer(ctr containerModel) armadav1.Container {
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

func toConfigFile(file configFileModel) armadav1.ConfigFileMount {
	return armadav1.ConfigFileMount{
		Name:      file.Name.ValueString(),
		MountPath: file.MountPath.ValueString(),
	}
}

func toVolumeMount(mount volumeMountModel) kcorev1.VolumeMount {
	return kcorev1.VolumeMount{
		Name:        mount.Name.ValueString(),
		MountPath:   mount.MountPath.ValueString(),
		SubPath:     mount.SubPath.ValueString(),
		SubPathExpr: mount.SubPathExpr.ValueString(),
	}
}

func toPort(p portModel) armadav1.Port {
	return armadav1.Port{
		Name:               p.Name.ValueString(),
		Policy:             agonesv1.PortPolicy(p.Policy.ValueString()),
		ContainerPort:      uint16(p.ContainerPort.ValueInt32()), //nolint:gosec // Keep it simple.
		Protocol:           kcorev1.Protocol(p.Protocol.ValueString()),
		ProtectionProtocol: p.ProtectionProtocol.ValueStringPointer(),
	}
}

func toResourceRequirements(res *resourcesModel) kcorev1.ResourceRequirements {
	limits := toResourceList(res.Limits)
	requests := toResourceList(res.Requests)

	req := kcorev1.ResourceRequirements{}
	if len(limits) > 0 {
		req.Limits = limits
	}
	if len(requests) > 0 {
		req.Requests = requests
	}
	return req
}

func toResourceList(spec *resourceSpecModel) kcorev1.ResourceList {
	res := kcorev1.ResourceList{}
	if spec == nil {
		return res
	}
	if q := toQuantity(spec.CPU); !q.IsZero() {
		res[kcorev1.ResourceCPU] = q
	}
	if q := toQuantity(spec.Memory); !q.IsZero() {
		res[kcorev1.ResourceMemory] = q
	}
	return res
}

func toQuantity(val types.String) resource.Quantity {
	if !conv.IsKnown(val) {
		return resource.Quantity{}
	}
	return resource.MustParse(val.ValueString()) // Pre-validation required.
}

func toTerminationGracePeriodSeconds(cfg *terminationConfigModel) *int64 {
	if cfg == nil || !conv.IsKnown(cfg.GracePeriodSeconds) {
		return nil
	}
	return cfg.GracePeriodSeconds.ValueInt64Pointer()
}

func toHealthChecks(hc *healthChecksModel) agonesv1.Health {
	if hc == nil {
		return agonesv1.Health{}
	}
	return agonesv1.Health{
		Disabled:            hc.Disabled.ValueBool(),
		InitialDelaySeconds: hc.InitialDelaySeconds.ValueInt32(),
		PeriodSeconds:       hc.PeriodSeconds.ValueInt32(),
		FailureThreshold:    hc.FailureThreshold.ValueInt32(),
	}
}

func toDeploymentStrategy(strat *strategyModel) appsv1.DeploymentStrategy {
	switch {
	case strat != nil && conv.IsKnown(strat.Recreate):
		return appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		}
	case strat == nil:
		return appsv1.DeploymentStrategy{
			Type:          appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{},
		}
	default:
		return appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxSurge:       toIntOrString(strat.RollingUpdate.MaxSurge),
				MaxUnavailable: toIntOrString(strat.RollingUpdate.MaxUnavailable),
			},
		}
	}
}

func toIntOrString(val types.String) *intstr.IntOrString {
	if !conv.IsKnown(val) {
		return nil
	}
	is := intstr.Parse(val.ValueString())
	return &is
}

func toFixedInterval(scaling *autoscalingModel) *armadav1.ArmadaFixInterval {
	if scaling == nil || !conv.IsKnown(scaling.FixedIntervalSeconds) {
		return &armadav1.ArmadaFixInterval{}
	}
	return &armadav1.ArmadaFixInterval{
		Seconds: scaling.FixedIntervalSeconds.ValueInt32(),
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

type autoscalingModel struct {
	FixedIntervalSeconds types.Int32 `tfsdk:"fixed_interval_seconds"`
}

func newAutoscalingModel(obj armadav1.ArmadaAutoscaling) *autoscalingModel {
	if obj.FixedInterval == nil || obj.FixedInterval.Seconds == 0 {
		return nil
	}
	return &autoscalingModel{
		FixedIntervalSeconds: types.Int32Value(obj.FixedInterval.Seconds),
	}
}

type replicaModel struct {
	RegionType  types.String `tfsdk:"region_type"`
	MinReplicas types.Int32  `tfsdk:"min_replicas"`
	MaxReplicas types.Int32  `tfsdk:"max_replicas"`
	BufferSize  types.Int32  `tfsdk:"buffer_size"`
}

func newReplicas(obj armadav1.ArmadaRegionType) replicaModel {
	return replicaModel{
		RegionType:  types.StringValue(obj.Name),
		MinReplicas: types.Int32Value(obj.MinReplicas),
		MaxReplicas: types.Int32Value(obj.MaxReplicas),
		BufferSize:  types.Int32Value(obj.BufferSize),
	}
}

type containerModel struct {
	Name         types.String       `tfsdk:"name"`
	Image        imageModel         `tfsdk:"image"`
	Command      []types.String     `tfsdk:"command"`
	Args         []types.String     `tfsdk:"args"`
	Resources    *resourcesModel    `tfsdk:"resources"`
	Envs         []core.EnvVarModel `tfsdk:"envs"`
	Ports        []portModel        `tfsdk:"ports"`
	VolumeMounts []volumeMountModel `tfsdk:"volume_mounts"`
	ConfigFiles  []configFileModel  `tfsdk:"config_files"`
}

func newContainer(obj armadav1.Container) containerModel {
	return containerModel{
		Name: types.StringValue(obj.Name),
		Image: imageModel{
			Name:   types.StringValue(obj.Image),
			Branch: types.StringValue(obj.Branch),
		},
		Command:      conv.ForEachSliceItem(obj.Command, types.StringValue),
		Args:         conv.ForEachSliceItem(obj.Args, types.StringValue),
		Resources:    newResourcesModel(obj.Resources),
		Envs:         conv.ForEachSliceItem(obj.Env, core.NewEnvVarModel),
		Ports:        conv.ForEachSliceItem(obj.Ports, newPortModel),
		VolumeMounts: conv.ForEachSliceItem(obj.VolumeMounts, newVolumeMountModel),
		ConfigFiles:  conv.ForEachSliceItem(obj.ConfigFiles, newConfigFileModel),
	}
}

type imageModel struct {
	Name   types.String `tfsdk:"name"`
	Branch types.String `tfsdk:"branch"`
}

type resourcesModel struct {
	Limits   *resourceSpecModel `tfsdk:"limits"`
	Requests *resourceSpecModel `tfsdk:"requests"`
}

func newResourcesModel(obj kcorev1.ResourceRequirements) *resourcesModel {
	res := &resourcesModel{}
	if obj.Limits != nil {
		res.Limits = newResourceSpecModel(obj.Limits)
	}
	if obj.Requests != nil {
		res.Requests = newResourceSpecModel(obj.Requests)
	}
	return res
}

type resourceSpecModel struct {
	CPU    types.String `tfsdk:"cpu"`
	Memory types.String `tfsdk:"memory"`
}

func newResourceSpecModel(obj kcorev1.ResourceList) *resourceSpecModel {
	res := resourceSpecModel{}
	if obj.Cpu() != nil {
		res.CPU = types.StringValue(obj.Cpu().String())
	}
	if obj.Memory() != nil {
		res.Memory = types.StringValue(obj.Memory().String())
	}
	return &res
}

type portModel struct {
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	ContainerPort      types.Int32  `tfsdk:"container_port"`
	Policy             types.String `tfsdk:"policy"`
	ProtectionProtocol types.String `tfsdk:"protection_protocol"`
}

func newPortModel(obj armadav1.Port) portModel {
	return portModel{
		Name:               types.StringValue(obj.Name),
		Protocol:           types.StringValue(string(obj.Protocol)),
		ContainerPort:      types.Int32Value(int32(obj.ContainerPort)),
		Policy:             types.StringValue(string(obj.Policy)),
		ProtectionProtocol: conv.OptionalFunc(*obj.ProtectionProtocol, types.StringValue, types.StringNull),
	}
}

type volumeMountModel struct {
	Name        types.String `tfsdk:"name"`
	MountPath   types.String `tfsdk:"mount_path"`
	SubPath     types.String `tfsdk:"sub_path"`
	SubPathExpr types.String `tfsdk:"sub_path_expr"`
}

func newVolumeMountModel(obj kcorev1.VolumeMount) volumeMountModel {
	return volumeMountModel{
		Name:        types.StringValue(obj.Name),
		MountPath:   types.StringValue(obj.MountPath),
		SubPath:     conv.OptionalFunc(obj.SubPath, types.StringValue, types.StringNull),
		SubPathExpr: conv.OptionalFunc(obj.SubPathExpr, types.StringValue, types.StringNull),
	}
}

type configFileModel struct {
	Name      types.String `tfsdk:"name"`
	MountPath types.String `tfsdk:"mount_path"`
}

func newConfigFileModel(obj armadav1.ConfigFileMount) configFileModel {
	return configFileModel{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

type healthChecksModel struct {
	Disabled            types.Bool  `tfsdk:"disabled"`
	InitialDelaySeconds types.Int32 `tfsdk:"initial_delay_seconds"`
	PeriodSeconds       types.Int32 `tfsdk:"period_seconds"`
	FailureThreshold    types.Int32 `tfsdk:"failure_threshold"`
}

type terminationConfigModel struct {
	GracePeriodSeconds types.Int64 `tfsdk:"grace_period_seconds"`
}

type strategyModel struct {
	RollingUpdate *rollingUpdateModel `tfsdk:"rolling_update"`
	Recreate      types.Object        `tfsdk:"recreate"`
}

func newStrategyModel(obj appsv1.DeploymentStrategy) *strategyModel {
	switch obj.Type {
	case appsv1.RollingUpdateDeploymentStrategyType:
		if obj.RollingUpdate == nil || (obj.RollingUpdate.MaxSurge == nil && obj.RollingUpdate.MaxUnavailable == nil) {
			return nil
		}
		return &strategyModel{
			RollingUpdate: &rollingUpdateModel{
				MaxSurge:       conv.FromIntOrString(obj.RollingUpdate.MaxSurge),
				MaxUnavailable: conv.FromIntOrString(obj.RollingUpdate.MaxUnavailable),
			},
		}
	default:
		return &strategyModel{
			Recreate: types.ObjectValueMust(nil, nil),
		}
	}
}

type rollingUpdateModel struct {
	MaxSurge       types.String `tfsdk:"max_surge"`
	MaxUnavailable types.String `tfsdk:"max_unavailable"`
}

type volumeModel struct {
	Name     types.String   `tfsdk:"name"`
	EmptyDir *emptyDirModel `tfsdk:"empty_dir"`
}

func newVolumeModel(obj armadav1.Volume) volumeModel {
	vol := volumeModel{
		Name: types.StringValue(obj.Name),
	}
	if obj.SizeLimit != nil && !obj.SizeLimit.IsZero() {
		vol.EmptyDir = &emptyDirModel{
			SizeLimit: conv.OptionalFunc(obj.SizeLimit.String(), types.StringValue, types.StringNull),
		}
	}
	return vol
}

type emptyDirModel struct {
	SizeLimit types.String `tfsdk:"size_limit"`
}
