package mps

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kcorev1 "k8s.io/api/core/v1"
)

// ContainerModel is the terraform model for a container.
type ContainerModel struct {
	Name         types.String       `tfsdk:"name"`
	ImageRef     ImageRefModel      `tfsdk:"image_ref"`
	Command      []types.String     `tfsdk:"command"`
	Args         []types.String     `tfsdk:"args"`
	Resources    *ResourcesModel    `tfsdk:"resources"`
	Envs         []core.EnvVarModel `tfsdk:"envs"`
	Ports        []PortModel        `tfsdk:"ports"`
	VolumeMounts []VolumeMountModel `tfsdk:"volume_mounts"`
	ConfigFiles  []ConfigFileModel  `tfsdk:"config_files"`
	Secrets      []SecretMountModel `tfsdk:"secrets"`
}

// NewContainerForArmada converts the backend resource into the terraform model.
func NewContainerForArmada(obj armadav1.Container) ContainerModel {
	return ContainerModel{
		Name: types.StringValue(obj.Name),
		ImageRef: ImageRefModel{
			Name:   types.StringValue(obj.Image),
			Branch: types.StringValue(obj.Branch),
		},
		Command:      conv.ForEachSliceItem(obj.Command, types.StringValue),
		Args:         conv.ForEachSliceItem(obj.Args, types.StringValue),
		Resources:    newResourcesModel(obj.Resources),
		Envs:         conv.ForEachSliceItem(obj.Env, core.NewEnvVarModel),
		Ports:        conv.ForEachSliceItem(obj.Ports, newPortModelForArmada),
		VolumeMounts: conv.ForEachSliceItem(obj.VolumeMounts, newVolumeMountModel),
		ConfigFiles:  conv.ForEachSliceItem(obj.ConfigFiles, NewConfigFileModelForArmada),
		Secrets:      conv.ForEachSliceItem(obj.Secrets, NewSecretMountModelForArmada),
	}
}

// NewContainerForFormation converts the backend resource into the terraform model.
func NewContainerForFormation(obj formationv1.Container) ContainerModel {
	return ContainerModel{
		Name: types.StringValue(obj.Name),
		ImageRef: ImageRefModel{
			Name:   types.StringValue(obj.Image),
			Branch: types.StringValue(obj.Branch),
		},
		Command:      conv.ForEachSliceItem(obj.Command, types.StringValue),
		Args:         conv.ForEachSliceItem(obj.Args, types.StringValue),
		Resources:    newResourcesModel(obj.Resources),
		Envs:         conv.ForEachSliceItem(obj.Env, core.NewEnvVarModel),
		Ports:        conv.ForEachSliceItem(obj.Ports, newPortModelForFormation),
		VolumeMounts: conv.ForEachSliceItem(obj.VolumeMounts, newVolumeMountModel),
		ConfigFiles:  conv.ForEachSliceItem(obj.ConfigFiles, newConfigFileModelForFormation),
		Secrets:      conv.ForEachSliceItem(obj.Secrets, newSecretMountModelForFormation),
	}
}

// ToContainerForArmada converts the terraform model into the backend resource.
func ToContainerForArmada(ctr ContainerModel) armadav1.Container {
	return armadav1.Container{
		Name:         ctr.Name.ValueString(),
		Image:        ctr.ImageRef.Name.ValueString(),
		Branch:       ctr.ImageRef.Branch.ValueString(),
		Command:      conv.ForEachSliceItem(ctr.Command, func(v types.String) string { return v.ValueString() }),
		Args:         conv.ForEachSliceItem(ctr.Args, func(v types.String) string { return v.ValueString() }),
		Resources:    toResourceRequirements(ctr.Resources),
		Env:          conv.ForEachSliceItem(ctr.Envs, func(item core.EnvVarModel) corev1.EnvVar { return item.ToObject() }),
		Ports:        conv.ForEachSliceItem(ctr.Ports, toPortForArmada),
		VolumeMounts: conv.ForEachSliceItem(ctr.VolumeMounts, toVolumeMount),
		ConfigFiles:  conv.ForEachSliceItem(ctr.ConfigFiles, ToConfigFileForArmada),
		Secrets:      conv.ForEachSliceItem(ctr.Secrets, ToArmadaSecretMount),
	}
}

// ToContainerForFormation converts the terraform model into the backend resource.
func ToContainerForFormation(ctr ContainerModel) formationv1.Container {
	return formationv1.Container{
		Name:         ctr.Name.ValueString(),
		Image:        ctr.ImageRef.Name.ValueString(),
		Branch:       ctr.ImageRef.Branch.ValueString(),
		Command:      conv.ForEachSliceItem(ctr.Command, func(v types.String) string { return v.ValueString() }),
		Args:         conv.ForEachSliceItem(ctr.Args, func(v types.String) string { return v.ValueString() }),
		Resources:    toResourceRequirements(ctr.Resources),
		Env:          conv.ForEachSliceItem(ctr.Envs, func(item core.EnvVarModel) corev1.EnvVar { return item.ToObject() }),
		Ports:        conv.ForEachSliceItem(ctr.Ports, toPortForFormation),
		VolumeMounts: conv.ForEachSliceItem(ctr.VolumeMounts, toVolumeMount),
		ConfigFiles:  conv.ForEachSliceItem(ctr.ConfigFiles, toConfigFileForFormation),
		Secrets:      conv.ForEachSliceItem(ctr.Secrets, toFormationSecretMount),
	}
}

// ImageRefModel is the terraform model for a container image.
type ImageRefModel struct {
	Name   types.String `tfsdk:"name"`
	Branch types.String `tfsdk:"branch"`
}

// ResourcesModel is the terraform model for container resource requirements.
type ResourcesModel struct {
	Limits   *ResourceSpecModel `tfsdk:"limits"`
	Requests *ResourceSpecModel `tfsdk:"requests"`
}

func newResourcesModel(obj kcorev1.ResourceRequirements) *ResourcesModel {
	res := &ResourcesModel{}
	if obj.Limits != nil {
		res.Limits = newResourceSpecModel(obj.Limits)
	}
	if obj.Requests != nil {
		res.Requests = newResourceSpecModel(obj.Requests)
	}
	return res
}

func toResourceRequirements(res *ResourcesModel) kcorev1.ResourceRequirements {
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

// ResourceSpecModel is the terraform model for resource specifications.
type ResourceSpecModel struct {
	CPU    types.String `tfsdk:"cpu"`
	Memory types.String `tfsdk:"memory"`
}

func newResourceSpecModel(obj kcorev1.ResourceList) *ResourceSpecModel {
	res := ResourceSpecModel{}
	if obj.Cpu() != nil {
		res.CPU = types.StringValue(obj.Cpu().String())
	}
	if obj.Memory() != nil {
		res.Memory = types.StringValue(obj.Memory().String())
	}
	return &res
}

// toResourceList converts a ResourceSpecModel to a kcorev1.ResourceList.
//
// The resources are expected to be valid resources, ensured via validation, otherwise this panics.
func toResourceList(spec *ResourceSpecModel) kcorev1.ResourceList {
	res := kcorev1.ResourceList{}
	if spec == nil {
		return res
	}
	if q := conv.Quantity(spec.CPU); q != nil && !q.IsZero() {
		res[kcorev1.ResourceCPU] = *q
	}
	if q := conv.Quantity(spec.Memory); q != nil && !q.IsZero() {
		res[kcorev1.ResourceMemory] = *q
	}
	return res
}

// PortModel is the terraform model for a container port.
type PortModel struct {
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	ContainerPort      types.Int32  `tfsdk:"container_port"`
	Policy             types.String `tfsdk:"policy"`
	ProtectionProtocol types.String `tfsdk:"protection_protocol"`
}

func newPortModelForArmada(obj armadav1.Port) PortModel {
	prot := types.StringNull()
	if obj.ProtectionProtocol != nil {
		prot = conv.OptionalFunc(*obj.ProtectionProtocol, types.StringValue, types.StringNull)
	}

	return PortModel{
		Name:               types.StringValue(obj.Name),
		Protocol:           types.StringValue(string(obj.Protocol)),
		ContainerPort:      types.Int32Value(int32(obj.ContainerPort)),
		Policy:             types.StringValue(string(obj.Policy)),
		ProtectionProtocol: prot,
	}
}

func newPortModelForFormation(obj formationv1.Port) PortModel {
	prot := types.StringNull()
	if obj.ProtectionProtocol != nil {
		prot = conv.OptionalFunc(*obj.ProtectionProtocol, types.StringValue, types.StringNull)
	}

	return PortModel{
		Name:               types.StringValue(obj.Name),
		Protocol:           types.StringValue(string(obj.Protocol)),
		ContainerPort:      types.Int32Value(int32(obj.ContainerPort)),
		Policy:             types.StringValue(string(obj.Policy)),
		ProtectionProtocol: prot,
	}
}

func toPortForArmada(p PortModel) armadav1.Port {
	return armadav1.Port{
		Name:               p.Name.ValueString(),
		Policy:             agonesv1.PortPolicy(p.Policy.ValueString()),
		ContainerPort:      uint16(p.ContainerPort.ValueInt32()), //nolint:gosec // Keep it simple.
		Protocol:           kcorev1.Protocol(p.Protocol.ValueString()),
		ProtectionProtocol: p.ProtectionProtocol.ValueStringPointer(),
	}
}

func toPortForFormation(p PortModel) formationv1.Port {
	return formationv1.Port{
		Name:               p.Name.ValueString(),
		Policy:             agonesv1.PortPolicy(p.Policy.ValueString()),
		ContainerPort:      uint16(p.ContainerPort.ValueInt32()), //nolint:gosec // Keep it simple.
		Protocol:           kcorev1.Protocol(p.Protocol.ValueString()),
		ProtectionProtocol: p.ProtectionProtocol.ValueStringPointer(),
	}
}

// VolumeMountModel is the terraform model for a container volume mount.
type VolumeMountModel struct {
	Name        types.String `tfsdk:"name"`
	MountPath   types.String `tfsdk:"mount_path"`
	SubPath     types.String `tfsdk:"sub_path"`
	SubPathExpr types.String `tfsdk:"sub_path_expr"`
}

func newVolumeMountModel(obj kcorev1.VolumeMount) VolumeMountModel {
	return VolumeMountModel{
		Name:        types.StringValue(obj.Name),
		MountPath:   types.StringValue(obj.MountPath),
		SubPath:     conv.OptionalFunc(obj.SubPath, types.StringValue, types.StringNull),
		SubPathExpr: conv.OptionalFunc(obj.SubPathExpr, types.StringValue, types.StringNull),
	}
}

func toVolumeMount(mount VolumeMountModel) kcorev1.VolumeMount {
	return kcorev1.VolumeMount{
		Name:        mount.Name.ValueString(),
		MountPath:   mount.MountPath.ValueString(),
		SubPath:     mount.SubPath.ValueString(),
		SubPathExpr: mount.SubPathExpr.ValueString(),
	}
}

// ConfigFileModel is the terraform model for a container config file mount.
type ConfigFileModel struct {
	Name      types.String `tfsdk:"name"`
	MountPath types.String `tfsdk:"mount_path"`
}

func NewConfigFileModelForArmada(obj armadav1.ConfigFileMount) ConfigFileModel {
	return ConfigFileModel{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

func newConfigFileModelForFormation(obj formationv1.ConfigFileMount) ConfigFileModel {
	return ConfigFileModel{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

func ToConfigFileForArmada(file ConfigFileModel) armadav1.ConfigFileMount {
	return armadav1.ConfigFileMount{
		Name:      file.Name.ValueString(),
		MountPath: file.MountPath.ValueString(),
	}
}

func toConfigFileForFormation(file ConfigFileModel) formationv1.ConfigFileMount {
	return formationv1.ConfigFileMount{
		Name:      file.Name.ValueString(),
		MountPath: file.MountPath.ValueString(),
	}
}

// SecretMountModel is the terraform model for a container secret mount.
type SecretMountModel struct {
	Name      types.String `tfsdk:"name"`
	MountPath types.String `tfsdk:"mount_path"`
}

func NewSecretMountModelForArmada(obj armadav1.SecretMount) SecretMountModel {
	return SecretMountModel{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

func newSecretMountModelForFormation(obj formationv1.SecretMount) SecretMountModel {
	return SecretMountModel{
		Name:      types.StringValue(obj.Name),
		MountPath: types.StringValue(obj.MountPath),
	}
}

func ToArmadaSecretMount(secret SecretMountModel) armadav1.SecretMount {
	return armadav1.SecretMount{
		Name:      secret.Name.ValueString(),
		MountPath: secret.MountPath.ValueString(),
	}
}

func toFormationSecretMount(secret SecretMountModel) formationv1.SecretMount {
	return formationv1.SecretMount{
		Name:      secret.Name.ValueString(),
		MountPath: secret.MountPath.ValueString(),
	}
}
