package mps

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kcorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerModel is the terraform model for a container.
type ContainerModel struct {
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

// NewContainerForArmada converts the backend resource into the terraform model.
func NewContainerForArmada(obj armadav1.Container) ContainerModel {
	return ContainerModel{
		Name: types.StringValue(obj.Name),
		Image: imageModel{
			Name:   types.StringValue(obj.Image),
			Branch: types.StringValue(obj.Branch),
		},
		Command:      conv.ForEachSliceItem(obj.Command, types.StringValue),
		Args:         conv.ForEachSliceItem(obj.Args, types.StringValue),
		Resources:    newResourcesModel(obj.Resources),
		Envs:         conv.ForEachSliceItem(obj.Env, core.NewEnvVarModel),
		Ports:        conv.ForEachSliceItem(obj.Ports, newPortModelForArmada),
		VolumeMounts: conv.ForEachSliceItem(obj.VolumeMounts, newVolumeMountModel),
		ConfigFiles:  conv.ForEachSliceItem(obj.ConfigFiles, newConfigFileModelForArmada),
	}
}

// ToContainerForArmada converts the terraform model into the backend resource.
func ToContainerForArmada(ctr ContainerModel) armadav1.Container {
	return armadav1.Container{
		Name:         ctr.Name.ValueString(),
		Image:        ctr.Image.Name.ValueString(),
		Branch:       ctr.Image.Branch.ValueString(),
		Command:      conv.ForEachSliceItem(ctr.Command, func(v types.String) string { return v.ValueString() }),
		Args:         conv.ForEachSliceItem(ctr.Args, func(v types.String) string { return v.ValueString() }),
		Resources:    toResourceRequirements(ctr.Resources),
		Env:          conv.ForEachSliceItem(ctr.Envs, func(item core.EnvVarModel) corev1.EnvVar { return item.ToObject() }),
		Ports:        conv.ForEachSliceItem(ctr.Ports, toPortForArmada),
		VolumeMounts: conv.ForEachSliceItem(ctr.VolumeMounts, toVolumeMount),
		ConfigFiles:  conv.ForEachSliceItem(ctr.ConfigFiles, toConfigFile),
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

// toResourceList converts a resourceSpecModel to a kcorev1.ResourceList.
//
// The resources are expected to be valid resources, ensured via validation, otherwise this panics.
func toResourceList(spec *resourceSpecModel) kcorev1.ResourceList {
	res := kcorev1.ResourceList{}
	if spec == nil {
		return res
	}
	if q := resource.MustParse(spec.CPU.ValueString()); !q.IsZero() {
		res[kcorev1.ResourceCPU] = q
	}
	if q := resource.MustParse(spec.Memory.ValueString()); !q.IsZero() {
		res[kcorev1.ResourceMemory] = q
	}
	return res
}

type portModel struct {
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	ContainerPort      types.Int32  `tfsdk:"container_port"`
	Policy             types.String `tfsdk:"policy"`
	ProtectionProtocol types.String `tfsdk:"protection_protocol"`
}

func newPortModelForArmada(obj armadav1.Port) portModel {
	return portModel{
		Name:               types.StringValue(obj.Name),
		Protocol:           types.StringValue(string(obj.Protocol)),
		ContainerPort:      types.Int32Value(int32(obj.ContainerPort)),
		Policy:             types.StringValue(string(obj.Policy)),
		ProtectionProtocol: conv.OptionalFunc(*obj.ProtectionProtocol, types.StringValue, types.StringNull),
	}
}

func toPortForArmada(p portModel) armadav1.Port {
	return armadav1.Port{
		Name:               p.Name.ValueString(),
		Policy:             agonesv1.PortPolicy(p.Policy.ValueString()),
		ContainerPort:      uint16(p.ContainerPort.ValueInt32()), //nolint:gosec // Keep it simple.
		Protocol:           kcorev1.Protocol(p.Protocol.ValueString()),
		ProtectionProtocol: p.ProtectionProtocol.ValueStringPointer(),
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

func toVolumeMount(mount volumeMountModel) kcorev1.VolumeMount {
	return kcorev1.VolumeMount{
		Name:        mount.Name.ValueString(),
		MountPath:   mount.MountPath.ValueString(),
		SubPath:     mount.SubPath.ValueString(),
		SubPathExpr: mount.SubPathExpr.ValueString(),
	}
}

type configFileModel struct {
	Name      types.String `tfsdk:"name"`
	MountPath types.String `tfsdk:"mount_path"`
}

func newConfigFileModelForArmada(obj armadav1.ConfigFileMount) configFileModel {
	return configFileModel{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

func toConfigFile(file configFileModel) armadav1.ConfigFileMount {
	return armadav1.ConfigFileMount{
		Name:      file.Name.ValueString(),
		MountPath: file.MountPath.ValueString(),
	}
}
