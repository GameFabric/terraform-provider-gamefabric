package container

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ImageUpdaterTargetTypeArmadaSet = "armadaset"
	ImageUpdaterTargetTypeArmada    = "armada"
	ImageUpdaterTargetTypeFormation = "formation"
	ImageUpdaterTargetTypeVessel    = "vessel"
)

type imageUpdaterModel struct {
	ID     types.String            `tfsdk:"id"`
	Branch types.String            `tfsdk:"branch"`
	Image  types.String            `tfsdk:"image"`
	Target ImageUpdaterTargetModel `tfsdk:"target"`
}

func newImageUpdaterModel(obj *containerv1.ImageUpdater) imageUpdaterModel {
	model := imageUpdaterModel{
		ID:     types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Branch: types.StringValue(obj.Spec.Branch),
		Image:  types.StringValue(obj.Spec.ImageName),
		Target: NewImageUpdaterTargetModel(obj.Spec.TargetRef, obj.Environment),
	}
	return model
}

func (m imageUpdaterModel) ToObject(name string) *containerv1.ImageUpdater {
	obj := &containerv1.ImageUpdater{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Environment: m.Target.Environment.ValueString(),
		},
		Spec: containerv1.ImageUpdaterSpec{
			Branch:    m.Branch.ValueString(),
			ImageName: m.Image.ValueString(),
			TargetRef: m.Target.ToObject(),
		},
	}
	return obj
}

// ImageUpdaterTargetModel is the image updater target model.
type ImageUpdaterTargetModel struct {
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Environment types.String `tfsdk:"environment"`
}

// NewImageUpdaterTargetModel creates a new image updater target model from the target reference.
func NewImageUpdaterTargetModel(obj containerv1.TargetRef, env string) ImageUpdaterTargetModel {
	typ := "unknown"
	switch {
	case obj.APIVersion == armadav1.GroupVersion.String() && obj.Kind == "ArmadaSet":
		typ = ImageUpdaterTargetTypeArmadaSet
	case obj.APIVersion == armadav1.GroupVersion.String() && obj.Kind == "Armada":
		typ = ImageUpdaterTargetTypeArmada
	case obj.APIVersion == formationv1.GroupVersion.String() && obj.Kind == "Formation":
		typ = ImageUpdaterTargetTypeFormation
	case obj.APIVersion == formationv1.GroupVersion.String() && obj.Kind == "Vessel":
		typ = ImageUpdaterTargetTypeVessel
	}

	model := ImageUpdaterTargetModel{
		Type:        types.StringValue(typ),
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(env),
	}
	return model
}

// ToObject converts the model to an API object.
func (m ImageUpdaterTargetModel) ToObject() containerv1.TargetRef {
	obj := containerv1.TargetRef{
		Name: m.Name.ValueString(),
	}
	switch m.Type.ValueString() {
	case ImageUpdaterTargetTypeArmadaSet:
		obj.APIVersion = armadav1.GroupVersion.String()
		obj.Kind = "ArmadaSet"
	case ImageUpdaterTargetTypeArmada:
		obj.APIVersion = armadav1.GroupVersion.String()
		obj.Kind = "Armada"
	case ImageUpdaterTargetTypeFormation:
		obj.APIVersion = formationv1.GroupVersion.String()
		obj.Kind = "Formation"
	case ImageUpdaterTargetTypeVessel:
		obj.APIVersion = formationv1.GroupVersion.String()
		obj.Kind = "Vessel"
	}
	return obj
}
