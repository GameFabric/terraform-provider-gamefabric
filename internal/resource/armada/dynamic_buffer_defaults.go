package armada

import (
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func applyDynamicBufferDefaultsToArmadaModel(model *armadaModel) {
	if model == nil {
		return
	}
	applyDynamicBufferDefaultsToReplicas(model.Replicas)
}

func applyDynamicBufferDefaultsToArmadaSetModel(model *armadaSetModel) {
	if model == nil {
		return
	}
	for i := range model.Regions {
		applyDynamicBufferDefaultsToReplicas(model.Regions[i].Replicas)
	}
}

func applyDynamicBufferDefaultsToReplicas(replicas []replicaModel) {
	for i := range replicas {
		applyDynamicBufferDefaultsToModel(replicas[i].DynamicBuffer)
	}
}

func applyDynamicBufferDefaultsToModel(model *dynamicBufferModel) {
	if model == nil || !conv.IsKnown(model.MaxBufferUtilization) {
		return
	}

	percentage := model.MaxBufferUtilization.ValueInt32()
	if !conv.IsKnown(model.DynamicMaxBufferThreshold) {
		model.DynamicMaxBufferThreshold = types.Int32Value(derivedDynamicMaxBufferThreshold(percentage))
	}
	if !conv.IsKnown(model.DynamicMinBufferThreshold) {
		model.DynamicMinBufferThreshold = types.Int32Value(derivedDynamicMinBufferThreshold(percentage))
	}
}
