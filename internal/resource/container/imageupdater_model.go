package container

import (
	"context"
	"fmt"

	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Image updater target types.
const (
	// ImageUpdaterTargetTypeArmadaSet is the Armada Set target type.
	ImageUpdaterTargetTypeArmadaSet = "armadaset"

	// ImageUpdaterTargetTypeArmada is the Armada target type.
	ImageUpdaterTargetTypeArmada = "armada"

	// ImageUpdaterTargetTypeFormation is the Formation target type.
	ImageUpdaterTargetTypeFormation = "formation"

	// ImageUpdaterTargetTypeVessel is the Vessel target type.
	ImageUpdaterTargetTypeVessel = "vessel"
)

// ImageUpdaterTargetModel is the image updater target model.
type ImageUpdaterTargetModel struct {
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Environment types.String `tfsdk:"environment"`

	known bool
}

// SetUnknown sets the unknown state of the model.
func (m *ImageUpdaterTargetModel) SetUnknown(_ context.Context, unknown bool) error {
	m.known = !unknown
	return nil
}

// SetValue sets the value of the model.
func (m *ImageUpdaterTargetModel) SetValue(_ context.Context, v any) error {
	target, ok := v.(ImageUpdaterTargetModel)
	if !ok {
		return fmt.Errorf("unsupported value type %T for ImageUpdaterTargetModel", v)
	}
	*m = target
	m.known = true
	return nil
}

// GetUnknown gets the unknown state of the model.
func (m *ImageUpdaterTargetModel) GetUnknown(context.Context) bool {
	return !m.known
}

// GetValue gets the value of the model.
func (m *ImageUpdaterTargetModel) GetValue(ctx context.Context) any {
	typ, _ := m.Type.ToTerraformValue(ctx)
	name, _ := m.Name.ToTerraformValue(ctx)
	env, _ := m.Environment.ToTerraformValue(ctx)
	return map[string]tftypes.Value{
		"type":        typ,
		"name":        name,
		"environment": env,
	}
}

// NewImageUpdaterTargetModel creates a new image updater target model.
func NewImageUpdaterTargetModel(typ, name, env string) *ImageUpdaterTargetModel {
	return &ImageUpdaterTargetModel{
		Type:        types.StringValue(typ),
		Name:        types.StringValue(name),
		Environment: types.StringValue(env),
		known:       true,
	}
}

// NewImageUpdaterTargetModelFromTarget creates a new image updater target model from the target reference.
func NewImageUpdaterTargetModelFromTarget(obj containerv1.TargetRef, env string) ImageUpdaterTargetModel {
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
		known:       true,
	}
	return model
}

// ToObject converts the model to an API object.
func (m *ImageUpdaterTargetModel) ToObject() containerv1.TargetRef {
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
		Target: NewImageUpdaterTargetModelFromTarget(obj.Spec.TargetRef, obj.Environment),
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
