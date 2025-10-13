package armada

import (
	"context"
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	armadav1 "github.com/gamefabric/gf-core/pkg/api/armada/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/schema/armada"
	"github.com/hamba/pkg/v2/ptr"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kappsv1 "k8s.io/api/apps/v1"
	kcorev1 "k8s.io/api/core/v1"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ArmadaResource is the Armada resource.
type ArmadaResource struct { //nolint:revive // Be specific.
	clientSet clientset.Interface
}

// NewArmadaResource returns a new Armada resource.
func NewArmadaResource() resource.Resource {
	return &ArmadaResource{}
}

// Metadata sets the resource type name.
func (r *ArmadaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_armada_v1"
}

// Schema defines the schema for this data source.
func (r *ArmadaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "",
		MarkdownDescription: "",
		Attributes:          armada.ArmadaAttributes(),
		Version:             0,
	}
}

// Configure prepares the struct.
func (r *ArmadaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	procCtx, ok := req.ProviderData.(*provcontext.Context)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *provider.Context, got %T", req.ProviderData),
		)
		return
	}

	r.clientSet = procCtx.ClientSet
}

// Create creates the resource and sets the initial Terraform state on success.
func (r *ArmadaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan armada.ArmadaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj := armadaFromModel(plan)
	_, err := r.clientSet.ArmadaV1().Armadas(obj.Environment).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Error",
			fmt.Sprintf("Could not create Armada: %v", err),
		)
		return
	}

	var state armada.ArmadaModel
	// if err = r.conv.ToModel(outObj, &state); err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Conversion Error",
	// 		fmt.Sprintf("Could not convert Armada to model: %v", err),
	// 	)
	// 	return
	// }
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ArmadaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state armada.ArmadaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// resp.State.RemoveResource(ctx) // If not found.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the state on success.
func (r *ArmadaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state armada.ArmadaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the state on success.
func (r *ArmadaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state armada.ArmadaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource state.
func (r *ArmadaResource) ImportState(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse) {
	// TODO
}

func armadaFromModel(model armada.ArmadaModel) *armadav1.Armada {
	obj := &armadav1.Armada{
		ObjectMeta: metav1.ObjectMeta{
			Name:        model.Name.ValueString(),
			Environment: model.Environment.ValueString(),
		},
		Spec: armadav1.ArmadaSpec{
			Description: model.Description.ValueString(),
			Region:      model.Region.ValueString(),
			Distribution: conv.ForEachSliceItem(model.Replicas, func(item armada.ArmadaReplicaModel) armadav1.ArmadaRegionType {
				return armadav1.ArmadaRegionType{
					Name:        item.RegionType.ValueString(),
					MinReplicas: item.MinReplicas.ValueInt32(),
					MaxReplicas: item.MaxReplicas.ValueInt32(),
					BufferSize:  item.BufferSize.ValueInt32(),
				}
			}),
			Autoscaling: armadaAutoscalingFromModel(model.Autoscaling),
			Template:    fleetTemplateSpecFromModel(model.Template),
		},
	}
	return obj
}

func armadaAutoscalingFromModel(model *armada.ArmadaAutoscalingModel) armadav1.ArmadaAutoscaling {
	if model == nil {
		return armadav1.ArmadaAutoscaling{}
	}
	if model.FixedInterval.IsNull() || model.FixedInterval.IsUnknown() {
		return armadav1.ArmadaAutoscaling{}
	}

	return armadav1.ArmadaAutoscaling{
		FixedInterval: &armadav1.ArmadaFixInterval{
			Seconds: model.FixedInterval.ValueInt32(),
		},
	}
}

func fleetTemplateSpecFromModel(model *armada.FleetTemplateSpecModel) armadav1.FleetTemplateSpec {
	if model == nil {
		return armadav1.FleetTemplateSpec{}
	}

	obj := armadav1.FleetTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      conv.ForEachMapItem(model.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(model.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: armadav1.FleetSpec{
			Containers: containersFromModel(model.Containers),
			// 	Volumes:                       model.Volumes,
			GatewayPolicies: conv.ForEachSliceItem(model.GatewayPolicies, func(item types.String) string { return item.ValueString() }),
		},
	}

	if model.TerminationConfiguration != nil {
		if !model.TerminationConfiguration.GracePeriodSeconds.IsNull() && !model.TerminationConfiguration.GracePeriodSeconds.IsUnknown() {
			obj.Spec.TerminationGracePeriodSeconds = ptr.Of(model.TerminationConfiguration.GracePeriodSeconds.ValueInt64())
		}
	}

	if model.Strategy != nil {
		obj.Spec.Strategy = kappsv1.DeploymentStrategy{
			Type: kappsv1.RecreateDeploymentStrategyType,
		}
		if model.Strategy.RollingUpdate != nil {
			obj.Spec.Strategy.Type = kappsv1.RollingUpdateDeploymentStrategyType
			obj.Spec.Strategy.RollingUpdate = &kappsv1.RollingUpdateDeployment{
				MaxUnavailable: intOrStrFromModel(model.Strategy.RollingUpdate.MaxUnavailable),
				MaxSurge:       intOrStrFromModel(model.Strategy.RollingUpdate.MaxSurge),
			}
		}
	}

	if model.HealthChecks != nil {
		obj.Spec.Health = agonesv1.Health{
			Disabled:            model.HealthChecks.Disabled.ValueBool(),
			PeriodSeconds:       model.HealthChecks.PeriodSeconds.ValueInt32(),
			FailureThreshold:    model.HealthChecks.FailureThreshold.ValueInt32(),
			InitialDelaySeconds: model.HealthChecks.InitialDelaySeconds.ValueInt32(),
		}
	}

	return obj
}

func containersFromModel(model []armada.ContainerModel) []armadav1.Container {
	if model == nil {
		return nil
	}

	containers := make([]armadav1.Container, len(model))
	for i, item := range model {
		containers[i] = armadav1.Container{
			Name:   item.Name.ValueString(),
			Branch: item.Branch.ValueString(),
			Image:  item.Image.ValueString(),
			Ports: conv.ForEachSliceItem(item.Ports, func(item armada.PortModel) armadav1.Port {
				port := armadav1.Port{
					Name:          item.Name.ValueString(),
					Policy:        agonesv1.PortPolicy(item.Policy.ValueString()),
					ContainerPort: uint16(item.ContainerPort.ValueInt32()), //nolint:gosec
					Protocol:      kcorev1.Protocol(item.Protocol.ValueString()),
				}
				if !item.ProtectionProtocol.IsNull() && !item.ProtectionProtocol.IsUnknown() {
					port.ProtectionProtocol = item.ProtectionProtocol.ValueStringPointer()
				}
				return port
			}),
			Command: conv.ForEachSliceItem(item.Command, func(item types.String) string { return item.ValueString() }),
			Args:    conv.ForEachSliceItem(item.Args, func(item types.String) string { return item.ValueString() }),
			Env:     envVarsFromModel(item.Env),
			ConfigFiles: conv.ForEachSliceItem(item.ConfigFiles, func(item armada.ConfigFileMountModel) armadav1.ConfigFileMount {
				return armadav1.ConfigFileMount{
					Name:      item.Name.ValueString(),
					MountPath: item.MountPath.ValueString(),
				}
			}),
			VolumeMounts: conv.ForEachSliceItem(item.VolumeMounts, func(item armada.VolumeMountModel) kcorev1.VolumeMount {
				return kcorev1.VolumeMount{
					Name:      item.Name.ValueString(),
					MountPath: item.MountPath.ValueString(),
					ReadOnly:  item.ReadOnly.ValueBool(),
				}
			}),
		}
		if item.Resources != nil {
			containers[i].Resources = kcorev1.ResourceRequirements{
				Requests: toResourceList(conv.ForEachMapItem(item.Resources.Requests, quantityFromModel)),
				Limits:   toResourceList(conv.ForEachMapItem(item.Resources.Limits, quantityFromModel)),
				Claims: conv.ForEachSliceItem(item.Resources.Claims, func(item armada.ResourceClaimModel) kcorev1.ResourceClaim {
					return kcorev1.ResourceClaim{
						Name:    item.Name.ValueString(),
						Request: item.Request.ValueString(),
					}
				}),
			}
		}
	}
	return containers
}

func envVarsFromModel(model []core.EnvVarModel) []corev1.EnvVar {
	if model == nil {
		return nil
	}
	envVars := make([]corev1.EnvVar, len(model))
	for i, item := range model {
		envVars[i] = item.ToObject()
	}
	return envVars
}

func intOrStrToModel(i *intstr.IntOrString) types.String {
	if i == nil {
		return types.StringNull()
	}
	return types.StringValue(i.String())
}

func intOrStrFromModel(v types.String) *intstr.IntOrString {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	is := intstr.Parse(v.ValueString())
	return &is
}

func quantityToModel(q kresource.Quantity) attr.Value {
	return types.StringValue(q.String())
}

func quantityFromModel(v types.String) kresource.Quantity {
	if v.IsNull() || v.IsUnknown() {
		return kresource.Quantity{}
	}
	q, err := kresource.ParseQuantity(v.ValueString())
	if err != nil {
		panic(fmt.Errorf("parsing quantity: %w", err))
	}
	return q
}

func toResourceList(m map[string]kresource.Quantity) map[kcorev1.ResourceName]kresource.Quantity {
	resourceList := make(kcorev1.ResourceList, len(m))
	for key, value := range m {
		resourceList[kcorev1.ResourceName(key)] = value
	}
	return resourceList
}
