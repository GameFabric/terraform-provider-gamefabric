package armada

import (
	"maps"
	"slices"

	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type armadaSetModel struct {
	ID                    types.String                       `tfsdk:"id"`
	Name                  types.String                       `tfsdk:"name"`
	Environment           types.String                       `tfsdk:"environment"`
	Description           types.String                       `tfsdk:"description"`
	Labels                map[string]types.String            `tfsdk:"labels"`
	Annotations           map[string]types.String            `tfsdk:"annotations"`
	Autoscaling           *autoscalingModel                  `tfsdk:"autoscaling"`
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

func newArmadaSetModel(obj *armadav1.ArmadaSet) armadaSetModel {
	return armadaSetModel{
		ID:                    types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Description:           conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           newAnnotations(obj.Annotations),
		Autoscaling:           newAutoscalingModel(obj.Spec.Autoscaling),
		Regions:               newRegionModels(obj.Spec),
		GameServerLabels:      conv.ForEachMapItem(obj.Spec.Template.Labels, types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, mps.NewContainerForArmada),
		HealthChecks:          mps.NewHealthChecks(obj.Spec.Template.Spec.Health),
		TerminationConfig:     newTerminationConfig(obj.Spec.Template.Spec.TerminationGracePeriodSeconds),
		Strategy:              newStrategyModel(obj.Spec.Template.Spec.Strategy),
		Volumes:               conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolumeModel),
		GatewayPolicies:       conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled:      newProfilingEnabled(obj.Annotations),
		ImageUpdaterTarget:    container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeArmada, obj.Name, obj.Environment),
	}
}

func newRegionModels(spec armadav1.ArmadaSetSpec) []regionModel {
	regs := make(map[string]regionModel, len(spec.Armadas))
	for _, val := range spec.Armadas {
		reg := regs[val.Region]

		reg.Name = types.StringValue(val.Region)
		reg.Replicas = conv.ForEachSliceItem(val.Distribution, newReplicas)

		regs[val.Region] = reg
	}
	for _, val := range spec.Override {
		reg := regs[val.Region]

		reg.Name = types.StringValue(val.Region)
		reg.Envs = conv.ForEachSliceItem(val.Env, core.NewEnvVarModel)
		reg.Labels = conv.ForEachMapItem(val.Labels, types.StringValue)

		regs[val.Region] = reg
	}

	keys := slices.Collect(maps.Keys(regs))
	slices.Sort(keys)

	res := make([]regionModel, 0, len(regs))
	for _, key := range keys {
		res = append(res, regs[key])
	}
	return res
}

type regionModel struct {
	Name     types.String            `tfsdk:"name"`
	Replicas []replicaModel          `tfsdk:"replicas"`
	Envs     []core.EnvVarModel      `tfsdk:"envs"`
	Labels   map[string]types.String `tfsdk:"labels"`
}

func (m armadaSetModel) ToObject() *armadav1.ArmadaSet {
	return &armadav1.ArmadaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(
				toAnnotations(m.Annotations, m.ProfilingEnabled),
				func(v types.String) string { return v.ValueString() },
			),
		},
		Spec: armadav1.ArmadaSetSpec{
			Description: m.Description.ValueString(),
			Armadas:     conv.ForEachSliceItem(m.Regions, toArmadaTemplate),
			Override:    conv.ForEachSliceItem(m.Regions, toArmadaOverride),
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

func toArmadaTemplate(reg regionModel) armadav1.ArmadaTemplate {
	return armadav1.ArmadaTemplate{
		Region: reg.Name.ValueString(),
		Distribution: conv.ForEachSliceItem(reg.Replicas, func(r replicaModel) armadav1.ArmadaRegionType {
			return armadav1.ArmadaRegionType{
				Name:        r.RegionType.ValueString(),
				MinReplicas: r.MinReplicas.ValueInt32(),
				MaxReplicas: r.MaxReplicas.ValueInt32(),
				BufferSize:  r.BufferSize.ValueInt32(),
			}
		}),
	}
}

func toArmadaOverride(reg regionModel) armadav1.ArmadaOverride {
	return armadav1.ArmadaOverride{
		Region: reg.Name.ValueString(),
		Env:    conv.ForEachSliceItem(reg.Envs, func(item core.EnvVarModel) corev1.EnvVar { return item.ToObject() }),
		Labels: conv.ForEachMapItem(reg.Labels, func(v types.String) string { return v.ValueString() }),
	}
}
