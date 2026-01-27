package formation

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/mps"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const profilingKey = "g8c.io/profiling"

type vesselModel struct {
	ID                    types.String                       `tfsdk:"id"`
	Name                  types.String                       `tfsdk:"name"`
	Environment           types.String                       `tfsdk:"environment"`
	Region                types.String                       `tfsdk:"region"`
	Description           types.String                       `tfsdk:"description"`
	Suspend               types.Bool                         `tfsdk:"suspend"`
	Labels                map[string]types.String            `tfsdk:"labels"`
	Annotations           map[string]types.String            `tfsdk:"annotations"`
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

func newVesselModel(obj *formationv1.Vessel) vesselModel {
	return vesselModel{
		ID:                    types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:                  types.StringValue(obj.Name),
		Environment:           types.StringValue(obj.Environment),
		Region:                types.StringValue(obj.Spec.Region),
		Description:           conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Suspend:               conv.OptionalFunc(obj.Spec.Suspend, types.BoolPointerValue, types.BoolNull),
		Labels:                conv.ForEachMapItem(obj.Labels, types.StringValue),
		Annotations:           conv.ForEachMapItem(obj.Annotations, types.StringValue),
		GameServerLabels:      conv.ForEachMapItem(conv.MapWithoutKey(obj.Spec.Template.Labels, profilingKey), types.StringValue),
		GameServerAnnotations: conv.ForEachMapItem(obj.Spec.Template.Annotations, types.StringValue),
		Containers:            conv.ForEachSliceItem(obj.Spec.Template.Spec.Containers, mps.NewContainerForFormation),
		HealthChecks:          mps.NewHealthChecks(obj.Spec.Template.Spec.Health),
		TerminationConfig:     newTerminationConfig(obj.Spec.Template.Spec.TerminationGracePeriodSeconds, obj.Spec.TerminationGracePeriods),
		Volumes:               conv.ForEachSliceItem(obj.Spec.Template.Spec.Volumes, newVolumeModel),
		GatewayPolicies:       conv.ForEachSliceItem(obj.Spec.Template.Spec.GatewayPolicies, types.StringValue),
		ProfilingEnabled:      conv.BoolFromMapKey(obj.Spec.Template.Labels, profilingKey, types.BoolValue(false)),
		ImageUpdaterTarget:    container.NewImageUpdaterTargetModel(container.ImageUpdaterTargetTypeVessel, obj.Name, obj.Environment),
	}
}

func (m vesselModel) ToObject() *formationv1.Vessel {
	return &formationv1.Vessel{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(v types.String) string { return v.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(v types.String) string { return v.ValueString() }),
		},
		Spec: formationv1.VesselSpec{
			Description: m.Description.ValueString(),
			Region:      m.Region.ValueString(),
			Suspend:     m.Suspend.ValueBoolPointer(),
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
					TerminationGracePeriodSeconds: toTerminationGracePeriodSeconds(m.TerminationConfig),
					Volumes:                       conv.ForEachSliceItem(m.Volumes, toVolume),
				},
			},
			TerminationGracePeriods: toTerminationGracePeriods(m.TerminationConfig),
		},
	}
}

// GetContainers returns the containers from the vessel model.
// Implements the mps.ContainerProvider interface.
func (m vesselModel) GetContainers() []mps.ContainerModel {
	return m.Containers
}

func toTerminationGracePeriods(term *terminationConfigModel) *formationv1.VesselTerminationGracePeriods {
	if term == nil || !(conv.IsKnown(term.Maintenance) || conv.IsKnown(term.SpecChange) || conv.IsKnown(term.UserInitiated)) { //nolint:staticcheck // This is simpler.
		return nil
	}
	return &formationv1.VesselTerminationGracePeriods{
		Maintenance:   term.Maintenance.ValueInt64Pointer(),
		SpecChange:    term.SpecChange.ValueInt64Pointer(),
		UserInitiated: term.UserInitiated.ValueInt64Pointer(),
	}
}

type terminationConfigModel struct {
	GracePeriod   types.Int64 `tfsdk:"grace_period_seconds"`
	Maintenance   types.Int64 `tfsdk:"maintenance_seconds"`
	SpecChange    types.Int64 `tfsdk:"spec_change_seconds"`
	UserInitiated types.Int64 `tfsdk:"user_initiated_seconds"`
}

func newTerminationConfig(gracePeriod *int64, periods *formationv1.VesselTerminationGracePeriods) *terminationConfigModel {
	if gracePeriod == nil && periods == nil {
		return nil
	}

	term := &terminationConfigModel{
		GracePeriod: conv.OptionalFunc(gracePeriod, types.Int64PointerValue, types.Int64Null),
	}
	if periods != nil {
		term.Maintenance = conv.OptionalFunc(periods.Maintenance, types.Int64PointerValue, types.Int64Null)
		term.SpecChange = conv.OptionalFunc(periods.SpecChange, types.Int64PointerValue, types.Int64Null)
		term.UserInitiated = conv.OptionalFunc(periods.UserInitiated, types.Int64PointerValue, types.Int64Null)
	}
	return term
}

func toTerminationGracePeriodSeconds(cfg *terminationConfigModel) *int64 {
	if cfg == nil || !conv.IsKnown(cfg.GracePeriod) {
		return nil
	}
	return cfg.GracePeriod.ValueInt64Pointer()
}

type volumeModel struct {
	Name       types.String     `tfsdk:"name"`
	EmptyDir   *emptyDirModel   `tfsdk:"empty_dir"`
	Persistent *persistentModel `tfsdk:"persistent"`
}

func newVolumeModel(obj formationv1.Volume) volumeModel {
	vol := volumeModel{
		Name: types.StringValue(obj.Name),
	}
	switch obj.Type {
	case formationv1.VolumeTypePersistent:
		vol.Persistent = &persistentModel{
			VolumeName: types.StringValue(obj.Persistent.VolumeName),
		}
	default:
		vol.EmptyDir = &emptyDirModel{
			SizeLimit: conv.OptionalFunc(obj.EmptyDir.SizeLimit.String(), types.StringValue, types.StringNull),
		}
	}
	return vol
}

func toVolume(vol volumeModel) formationv1.Volume {
	res := formationv1.Volume{
		Name: vol.Name.ValueString(),
	}
	switch {
	case vol.Persistent != nil:
		res.Type = formationv1.VolumeTypePersistent
		res.Persistent = &formationv1.PersistentVolumeSource{
			VolumeName: vol.Persistent.VolumeName.ValueString(),
		}
	default:
		res.Type = formationv1.VolumeTypeEmptyDir
		res.EmptyDir = &formationv1.EmptyDirVolumeSource{
			SizeLimit: conv.Quantity(vol.EmptyDir.SizeLimit),
		}
	}
	return res
}

type emptyDirModel struct {
	SizeLimit types.String `tfsdk:"size_limit"`
}

type persistentModel struct {
	VolumeName types.String `tfsdk:"volume_name"`
}
