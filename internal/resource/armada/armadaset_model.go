package armada

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type armadaSetModel struct {
	ID                    types.String                       `tfsdk:"id"`
	Name                  types.String                       `tfsdk:"name"`
	Environment           types.String                       `tfsdk:"environment"`
	Description           types.String                       `tfsdk:"description"`
	Labels                map[string]types.String            `tfsdk:"labels"`
	Annotations           map[string]types.String            `tfsdk:"annotations"`
	Autoscaling           *armadaSetAutoscalingModel         `tfsdk:"autoscaling"`
	Regions               []regionModel                      `tfsdk:"regions"`
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

func newArmadaSetModel(obj *armadav1.ArmadaSet, as *armadaSetAutoscalingModel) armadaSetModel {
	return armadaSetModel{
		ID:                    types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Description:           conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           conv.ForEachMapItem(obj.Annotations, types.StringValue),
		Autoscaling:           newArmadaSetAutoscalingModel(obj.Spec.Autoscaling, as),
		Regions:               newRegionModels(obj.Spec),
		GameServerLabels:      conv.ForEachMapItem(conv.MapWithoutKey(obj.Spec.Template.Labels, profilingKey), types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, mps.NewContainerForArmada),
		HealthChecks:          mps.NewHealthChecks(obj.Spec.Template.Spec.Health),
		TerminationConfig:     newTerminationConfig(obj.Spec.Template.Spec.TerminationGracePeriodSeconds),
		Strategy:              newStrategyModel(obj.Spec.Template.Spec.Strategy),
		Volumes:               conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolumeModel),
		GatewayPolicies:       conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled:      conv.BoolFromMapKey(obj.Spec.Template.Labels, profilingKey, types.BoolValue(false)),
		ImageUpdaterTarget:    container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeArmadaSet, obj.Name, obj.Environment),
	}
}

func (m armadaSetModel) ToObject() *armadav1.ArmadaSet {
	return &armadav1.ArmadaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(v types.String) string { return v.ValueString() }),
		},
		Spec: armadav1.ArmadaSetSpec{
			Description: m.Description.ValueString(),
			Armadas: conv.ForEachSliceItem(m.Regions, func(reg regionModel) armadav1.ArmadaTemplate {
				return toArmadaTemplate(reg, m.Autoscaling)
			}),
			Override: conv.ForEachSliceItem(m.Regions, toArmadaOverride),
			Autoscaling: armadav1.ArmadaSetAutoscaling{
				FixedInterval: toArmadaSetFixedInterval(m.Autoscaling),
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

type armadaSetAutoscalingModel struct {
	FixedIntervalSeconds types.Int32 `tfsdk:"fixed_interval_seconds"`

	// Computed from the ArmadaSet Regions.
	ScaleToZero *scaleToZeroModel `tfsdk:"scale_to_zero"`
}

func toArmadaSetFixedInterval(scaling *armadaSetAutoscalingModel) *armadav1.ArmadaFixInterval {
	if scaling == nil || !conv.IsKnown(scaling.FixedIntervalSeconds) {
		return nil
	}
	return &armadav1.ArmadaFixInterval{
		Seconds: scaling.FixedIntervalSeconds.ValueInt32(),
	}
}

func newArmadaSetAutoscalingModel(obj armadav1.ArmadaSetAutoscaling, global *armadaSetAutoscalingModel) *armadaSetAutoscalingModel {
	if (global == nil || global.ScaleToZero == nil) && (obj.FixedInterval == nil || obj.FixedInterval.Seconds <= 0) {
		return nil
	}
	as := &armadaSetAutoscalingModel{
		FixedIntervalSeconds: types.Int32Value(obj.FixedInterval.Seconds),
	}
	if global != nil && global.ScaleToZero != nil && conv.IsKnown(global.ScaleToZero.ScaleUpUtilization) {
		// Keep the global scale to zero settings from the original state.
		// It has no relation to the object and would be lost otherwise.
		as.ScaleToZero = global.ScaleToZero
	}
	return as
}

func newRegionModels(spec armadav1.ArmadaSetSpec) []regionModel {
	regs := make([]regionModel, 0, len(spec.Armadas))
	regIdx := make(map[string]int, len(spec.Armadas))
	for _, arm := range spec.Armadas {
		reg := regionModel{
			Name:        types.StringValue(arm.Region),
			Replicas:    conv.ForEachSliceItem(arm.Distribution, newReplicas),
			Autoscaling: newArmadaTemplateAutoscaling(arm.Autoscaling),
		}
		regs = append(regs, reg)
		regIdx[arm.Region] = len(regs) - 1
	}

	for _, val := range spec.Override {
		idx, ok := regIdx[val.Region]
		if !ok {
			continue
		}

		reg := regs[idx]
		reg.Envs = conv.ForEachSliceItem(val.Env, core.NewEnvVarModel)
		reg.GameServerLabels = conv.ForEachMapItem(val.Labels, types.StringValue)
		reg.ConfigFiles = conv.ForEachSliceItem(val.ConfigFiles, mps.NewConfigFileModelForArmada)
		reg.Secrets = conv.ForEachSliceItem(val.Secrets, mps.NewSecretMountModelForArmada)
		regs[idx] = reg
	}
	return regs
}

func newArmadaTemplateAutoscaling(as armadav1.ArmadaTemplateAutoscaling) *armadaTemplateAutoscaling {
	if as.ScaleToZero == nil || (as.ScaleToZero.ScaleDownUtilization.IntValue() == 0 && as.ScaleToZero.ScaleUpUtilization.IntValue() == 0) {
		return nil
	}

	return &armadaTemplateAutoscaling{
		ScaleToZero: &scaleToZeroModel{
			ScaleDownUtilization: types.Int32Value(int32(as.ScaleToZero.ScaleDownUtilization.IntValue())),
			ScaleUpUtilization:   types.Int32Value(int32(as.ScaleToZero.ScaleUpUtilization.IntValue())),
		},
	}
}

type regionModel struct {
	Name             types.String               `tfsdk:"name"`
	Autoscaling      *armadaTemplateAutoscaling `tfsdk:"autoscaling"`
	Replicas         []replicaModel             `tfsdk:"replicas"`
	Envs             []core.EnvVarModel         `tfsdk:"envs"`
	GameServerLabels map[string]types.String    `tfsdk:"gameserver_labels"`
	ConfigFiles      []mps.ConfigFileModel      `tfsdk:"config_files"`
	Secrets          []mps.SecretMountModel     `tfsdk:"secrets"`
}

type armadaTemplateAutoscaling struct {
	ScaleToZero *scaleToZeroModel `tfsdk:"scale_to_zero"`
}

func toArmadaTemplate(reg regionModel, as *armadaSetAutoscalingModel) armadav1.ArmadaTemplate {
	return armadav1.ArmadaTemplate{
		Region:      reg.Name.ValueString(),
		Autoscaling: toArmadaTemplateAutoscaling(reg.Autoscaling, as),
		Distribution: conv.ForEachSliceItem(reg.Replicas, func(r replicaModel) armadav1.ArmadaRegionType {
			return armadav1.ArmadaRegionType{
				Name:          r.RegionType.ValueString(),
				MinReplicas:   r.MinReplicas.ValueInt32(),
				MaxReplicas:   r.MaxReplicas.ValueInt32(),
				BufferSize:    r.BufferSize.ValueInt32(),
				DynamicBuffer: toDynamicBuffer(r.DynamicBuffer),
			}
		}),
	}
}

func toArmadaTemplateAutoscaling(as *armadaTemplateAutoscaling, sharedAs *armadaSetAutoscalingModel) armadav1.ArmadaTemplateAutoscaling {
	if as == nil || as.ScaleToZero == nil || (as.ScaleToZero.ScaleDownUtilization.ValueInt32() == 0 && as.ScaleToZero.ScaleUpUtilization.ValueInt32() == 0) {
		// Check shared setting.
		if sharedAs == nil || sharedAs.ScaleToZero == nil || (sharedAs.ScaleToZero.ScaleDownUtilization.ValueInt32() == 0 && sharedAs.ScaleToZero.ScaleUpUtilization.ValueInt32() == 0) {
			return armadav1.ArmadaTemplateAutoscaling{}
		}
		return armadav1.ArmadaTemplateAutoscaling{
			ScaleToZero: &armadav1.ArmadaScaleToZero{
				ScaleDownUtilization: intstr.FromInt32(sharedAs.ScaleToZero.ScaleDownUtilization.ValueInt32()),
				ScaleUpUtilization:   intstr.FromInt32(sharedAs.ScaleToZero.ScaleUpUtilization.ValueInt32()),
			},
		}
	}
	return armadav1.ArmadaTemplateAutoscaling{
		ScaleToZero: &armadav1.ArmadaScaleToZero{
			ScaleDownUtilization: intstr.FromInt32(as.ScaleToZero.ScaleDownUtilization.ValueInt32()),
			ScaleUpUtilization:   intstr.FromInt32(as.ScaleToZero.ScaleUpUtilization.ValueInt32()),
		},
	}
}

func toArmadaOverride(reg regionModel) armadav1.ArmadaOverride {
	return armadav1.ArmadaOverride{
		Region:      reg.Name.ValueString(),
		Env:         conv.ForEachSliceItem(reg.Envs, func(item core.EnvVarModel) corev1.EnvVar { return item.ToObject() }),
		Labels:      conv.ForEachMapItem(reg.GameServerLabels, func(v types.String) string { return v.ValueString() }),
		ConfigFiles: conv.ForEachSliceItem(reg.ConfigFiles, mps.ToConfigFileForArmada),
		Secrets:     conv.ForEachSliceItem(reg.Secrets, mps.ToArmadaSecretMount),
	}
}
