package armada

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const profilingKey = "g8c.io/profiling"

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
	Containers            []mps.ContainerModel               `tfsdk:"containers"`
	HealthChecks          *mps.HealthChecksModel             `tfsdk:"health_checks"`
	TerminationConfig     *terminationConfigModel            `tfsdk:"termination_configuration"`
	Strategy              *strategyModel                     `tfsdk:"strategy"`
	Volumes               []volumeModel                      `tfsdk:"volumes"`
	GatewayPolicies       []types.String                     `tfsdk:"gateway_policies"`
	ProfilingEnabled      types.Bool                         `tfsdk:"profiling_enabled"`
	ImageUpdaterTarget    *container.ImageUpdaterTargetModel `tfsdk:"image_updater_target"`
}

func newArmadaModel(obj *armadav1.Armada) armadaModel {
	return armadaModel{
		ID:                    types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Description:           conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           conv.ForEachMapItem(obj.Annotations, types.StringValue),
		Autoscaling:           newAutoscalingModel(obj.Spec.Autoscaling),
		Region:                types.StringValue(obj.Spec.Region),
		Replicas:              conv.ForEachSliceItem(obj.Spec.Distribution, newReplicas),
		GameServerLabels:      conv.ForEachMapItem(conv.MapWithoutKey(obj.Spec.Template.Labels, profilingKey), types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, mps.NewContainerForArmada),
		HealthChecks:          mps.NewHealthChecks(obj.Spec.Template.Spec.Health),
		TerminationConfig:     newTerminationConfig(obj.Spec.Template.Spec.TerminationGracePeriodSeconds),
		Strategy:              newStrategyModel(obj.Spec.Template.Spec.Strategy),
		Volumes:               conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolumeModel),
		GatewayPolicies:       conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled:      conv.BoolFromMapKey(obj.Spec.Template.Labels, profilingKey, types.BoolValue(false)),
		ImageUpdaterTarget:    container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeArmada, obj.Name, obj.Environment),
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
					Labels: conv.ForEachMapItem(
						conv.MapWithBool(m.GameServerLabels, profilingKey, m.ProfilingEnabled),
						func(v types.String) string { return v.ValueString() },
					),
					Annotations: conv.ForEachMapItem(m.GameServerAnnotations, func(v types.String) string { return v.ValueString() }),
				},
				Spec: armadav1.FleetSpec{
					GatewayPolicies:               conv.ForEachSliceItem(m.GatewayPolicies, func(v types.String) string { return v.ValueString() }),
					Strategy:                      toStrategy(m.Strategy),
					Health:                        mps.ToHealthChecks(m.HealthChecks),
					Containers:                    conv.ForEachSliceItem(m.Containers, mps.ToContainerForArmada),
					TerminationGracePeriodSeconds: toTerminationGracePeriodSeconds(m.TerminationConfig),
					Volumes:                       conv.ForEachSliceItem(m.Volumes, toVolume),
				},
			},
		},
	}
}

func toStrategy(strat *strategyModel) appsv1.DeploymentStrategy {
	switch {
	case strat == nil:
		return appsv1.DeploymentStrategy{}
	case conv.IsKnown(strat.Recreate):
		return appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		}
	default:
		if strat.RollingUpdate == nil {
			return appsv1.DeploymentStrategy{}
		}
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

type autoscalingModel struct {
	FixedIntervalSeconds types.Int32 `tfsdk:"fixed_interval_seconds"`
}

func newAutoscalingModel(obj armadav1.ArmadaAutoscaling) *autoscalingModel {
	if obj.FixedInterval == nil {
		return nil
	}
	return &autoscalingModel{
		FixedIntervalSeconds: types.Int32Value(obj.FixedInterval.Seconds),
	}
}

func toFixedInterval(scaling *autoscalingModel) *armadav1.ArmadaFixInterval {
	if scaling == nil || !conv.IsKnown(scaling.FixedIntervalSeconds) {
		return nil
	}
	return &armadav1.ArmadaFixInterval{
		Seconds: scaling.FixedIntervalSeconds.ValueInt32(),
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

type terminationConfigModel struct {
	GracePeriodSeconds types.Int64 `tfsdk:"grace_period_seconds"`
}

func newTerminationConfig(seconds *int64) *terminationConfigModel {
	if seconds == nil {
		return nil
	}
	return &terminationConfigModel{
		GracePeriodSeconds: types.Int64Value(*seconds),
	}
}

func toTerminationGracePeriodSeconds(cfg *terminationConfigModel) *int64 {
	if cfg == nil || !conv.IsKnown(cfg.GracePeriodSeconds) {
		return nil
	}
	return cfg.GracePeriodSeconds.ValueInt64Pointer()
}

type strategyModel struct {
	RollingUpdate *rollingUpdateModel `tfsdk:"rolling_update"`
	Recreate      types.Object        `tfsdk:"recreate"`
}

func newStrategyModel(obj appsv1.DeploymentStrategy) *strategyModel {
	switch obj.Type {
	case appsv1.RollingUpdateDeploymentStrategyType:
		return &strategyModel{
			RollingUpdate: &rollingUpdateModel{
				MaxSurge:       conv.FromIntOrString(obj.RollingUpdate.MaxSurge),
				MaxUnavailable: conv.FromIntOrString(obj.RollingUpdate.MaxUnavailable),
			},
		}
	case appsv1.RecreateDeploymentStrategyType:
		return &strategyModel{
			Recreate: types.ObjectValueMust(nil, nil),
		}
	default:
		return nil
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

func toVolume(vol volumeModel) armadav1.Volume {
	res := armadav1.Volume{
		Name: vol.Name.ValueString(),
	}
	if vol.EmptyDir != nil {
		res.SizeLimit = conv.QuantityPointer(vol.EmptyDir.SizeLimit)
	}
	return res
}

type emptyDirModel struct {
	SizeLimit types.String `tfsdk:"size_limit"`
}
