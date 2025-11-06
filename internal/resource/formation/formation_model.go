package formation

import (
	"cmp"

	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	v1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type formationModel struct {
	ID                    types.String                       `tfsdk:"id"`
	Name                  types.String                       `tfsdk:"name"`
	Environment           types.String                       `tfsdk:"environment"`
	Description           types.String                       `tfsdk:"description"`
	Labels                map[string]types.String            `tfsdk:"labels"`
	Annotations           map[string]types.String            `tfsdk:"annotations"`
	VolumeTemplates       []VolumeTemplateModel              `tfsdk:"volume_templates"`
	Vessels               []VesselTemplateModel              `tfsdk:"vessels"`
	GameServerLabels      map[string]types.String            `tfsdk:"gameserver_labels"`
	GameServerAnnotations map[string]types.String            `tfsdk:"gameserver_annotations"`
	Containers            []mps.ContainerModel               `tfsdk:"containers"`
	HealthChecks          *mps.HealthChecksModel             `tfsdk:"health_checks"`
	TerminationConfig     *terminationConfigModel            `tfsdk:"termination_configuration"`
	Volumes               []volumeModel                      `tfsdk:"volumes"`
	GatewayPolicies       []types.String                     `tfsdk:"gateway_policies"`
	ProfilingEnabled      types.Bool                         `tfsdk:"profiling_enabled"`
	ImageUpdaterTarget    *container.ImageUpdaterTargetModel `tfsdk:"image_updater_target"`
}

func newFormationModel(obj *formationv1.Formation) formationModel {
	return formationModel{
		ID:                    types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Description:           conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           conv.ForEachMapItem(obj.Annotations, types.StringValue),
		VolumeTemplates:       conv.ForEachSliceItem(obj.Spec.VolumeTemplates, newVolumeTemplate),
		Vessels:               conv.ForEachSliceItem(obj.Spec.Vessels, newVesselTemplateModel),
		GameServerLabels:      conv.ForEachMapItem(conv.MapWithoutKey(obj.Spec.Template.Labels, profilingKey), types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, mps.NewContainerForFormation),
		HealthChecks:          mps.NewHealthChecks(obj.Spec.Template.Spec.Health),
		TerminationConfig:     newTerminationConfig(obj.Spec.Template.Spec.TerminationGracePeriodSeconds, obj.Spec.TerminationGracePeriods),
		Volumes:               conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolumeModel),
		GatewayPolicies:       conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled:      conv.BoolFromMapKey(obj.Spec.Template.Labels, profilingKey, types.BoolValue(false)),
		ImageUpdaterTarget:    container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeFormation, obj.Name, obj.Environment),
	}
}

func (m formationModel) ToObject() *formationv1.Formation {
	return &formationv1.Formation{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(v types.String) string { return v.ValueString() }),
		},
		Spec: formationv1.FormationSpec{
			Description: m.Description.ValueString(),
			Vessels:     conv.ForEachSliceItem(m.Vessels, toVesselTemplateModel),
			Template: formationv1.GameServerTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: conv.ForEachMapItem(
						conv.MapWithBool(m.GameServerLabels, profilingKey, m.ProfilingEnabled),
						func(v types.String) string { return v.ValueString() },
					),
					Annotations: conv.ForEachMapItem(m.GameServerAnnotations, func(v types.String) string { return v.ValueString() }),
				},
				Spec: formationv1.GameServerSpec{
					GatewayPolicies:               conv.ForEachSliceItem(m.GatewayPolicies, func(v types.String) string { return v.ValueString() }),
					Health:                        mps.ToHealthChecks(m.HealthChecks),
					Containers:                    conv.ForEachSliceItem(m.Containers, mps.ToContainerForFormation),
					TerminationGracePeriodSeconds: cmp.Or(m.TerminationConfig, &terminationConfigModel{}).GracePeriod.ValueInt64Pointer(),
					Volumes:                       conv.ForEachSliceItem(m.Volumes, toVolume),
				},
			},
			VolumeTemplates:         conv.ForEachSliceItem(m.VolumeTemplates, toVolumeTemplateModel),
			TerminationGracePeriods: toTerminationGracePeriods(m.TerminationConfig),
		},
	}
}

// VolumeTemplateModel represents a volume template within a formation.
type VolumeTemplateModel struct {
	Name            types.String `tfsdk:"name"`
	ReclaimPolicy   types.String `tfsdk:"reclaim_policy"`
	VolumeStoreName types.String `tfsdk:"volume_store_name"`
	Capacity        types.String `tfsdk:"capacity"`
}

func newVolumeTemplate(obj formationv1.VolumeTemplate) VolumeTemplateModel {
	return VolumeTemplateModel{
		Name:            types.StringValue(obj.Name),
		ReclaimPolicy:   types.StringValue(string(obj.Spec.ReclaimPolicy)),
		VolumeStoreName: types.StringValue(obj.Spec.VolumeStoreName),
		Capacity:        types.StringValue(obj.Spec.Capacity.String()),
	}
}

func toVolumeTemplateModel(m VolumeTemplateModel) formationv1.VolumeTemplate {
	return formationv1.VolumeTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name.ValueString(),
		},
		Spec: formationv1.VolumeTemplateSpec{
			VolumeSpec: v1beta1.VolumeSpec{
				VolumeStoreName: m.VolumeStoreName.ValueString(),
				Capacity:        *conv.Quantity(m.Capacity),
			},
			ReclaimPolicy: formationv1.VolumeTemplateReclaimPolicy(m.ReclaimPolicy.ValueString()),
		},
	}
}

// VesselTemplateModel represents a vessel template within a formation.
type VesselTemplateModel struct {
	Name        types.String         `tfsdk:"name"`
	Region      types.String         `tfsdk:"region"`
	Description types.String         `tfsdk:"description"`
	Suspend     types.Bool           `tfsdk:"suspend"`
	Override    *VesselOverrideModel `tfsdk:"override"`
}

func newVesselTemplateModel(obj formationv1.VesselTemplate) VesselTemplateModel {
	return VesselTemplateModel{
		Name:        types.StringValue(obj.Name),
		Region:      types.StringValue(obj.Region),
		Description: conv.OptionalFunc(obj.Description, types.StringValue, types.StringNull),
		Suspend:     conv.OptionalFunc(obj.Suspend, types.BoolPointerValue, types.BoolNull),
		Override:    newVesselOverrideModel(obj.Override),
	}
}

func toVesselTemplateModel(m VesselTemplateModel) formationv1.VesselTemplate {
	ovr := cmp.Or(m.Override, &VesselOverrideModel{})
	return formationv1.VesselTemplate{
		Name:        m.Name.ValueString(),
		Region:      m.Region.ValueString(),
		Description: m.Description.ValueString(),
		Suspend:     m.Suspend.ValueBoolPointer(),
		Override: formationv1.VesselOverride{
			Labels:     conv.ForEachMapItem(ovr.GameServerLabels, func(v types.String) string { return v.ValueString() }),
			Containers: conv.ForEachSliceItem(ovr.Containers, toContainerOverrideModel),
		},
	}
}

// VesselOverrideModel represents overrides for a vessel.
type VesselOverrideModel struct {
	GameServerLabels map[string]types.String        `tfsdk:"gameserver_labels"`
	Containers       []VesselContainerOverrideModel `tfsdk:"containers"`
}

func newVesselOverrideModel(obj formationv1.VesselOverride) *VesselOverrideModel {
	if len(obj.Labels) == 0 && len(obj.Containers) == 0 {
		return nil
	}
	return &VesselOverrideModel{
		GameServerLabels: conv.ForEachMapItem(obj.Labels, types.StringValue),
		Containers:       conv.ForEachSliceItem(obj.Containers, newContainerOverrideModel),
	}
}

// VesselContainerOverrideModel represents overrides for a container within a vessel.
type VesselContainerOverrideModel struct {
	Command []types.String     `tfsdk:"command"`
	Args    []types.String     `tfsdk:"args"`
	Envs    []core.EnvVarModel `tfsdk:"envs"`
}

func newContainerOverrideModel(obj formationv1.VesselContainerOverride) VesselContainerOverrideModel {
	return VesselContainerOverrideModel{
		Command: conv.ForEachSliceItem(obj.Command, types.StringValue),
		Args:    conv.ForEachSliceItem(obj.Args, types.StringValue),
		Envs:    conv.ForEachSliceItem(obj.Env, core.NewEnvVarModel),
	}
}

func toContainerOverrideModel(m VesselContainerOverrideModel) formationv1.VesselContainerOverride {
	return formationv1.VesselContainerOverride{
		Command: conv.ForEachSliceItem(m.Command, func(v types.String) string { return v.ValueString() }),
		Args:    conv.ForEachSliceItem(m.Args, func(v types.String) string { return v.ValueString() }),
		Env:     conv.ForEachSliceItem(m.Envs, func(m core.EnvVarModel) v1.EnvVar { return m.ToObject() }),
	}
}
