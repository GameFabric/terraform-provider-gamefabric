package provisioning

import (
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type allocatorModel struct {
	Name                  types.String   `tfsdk:"name"`
	Region                types.String   `tfsdk:"region"`
	RateLimitQPS          types.Int64    `tfsdk:"rate_limit_qps"`
	RateLimitBurst        types.Int64    `tfsdk:"rate_limit_burst"`
	AllocationURL         types.String   `tfsdk:"allocation_url"`
	AllocationActiveToken types.String   `tfsdk:"allocation_active_token"`
	AllocationTokens      []types.String `tfsdk:"allocation_tokens"`
	RegistryURL           types.String   `tfsdk:"registry_url"`
	RegistryActiveToken   types.String   `tfsdk:"registry_active_token"`
	RegistryTokens        []types.String `tfsdk:"registry_tokens"`
}

func newAllocatorModel(obj *provisioningv1beta1.Allocator) allocatorModel {
	return allocatorModel{
		Name:                  types.StringValue(obj.Name),
		Region:                types.StringValue(obj.Spec.Region),
		RateLimitQPS:          conv.OptionalFunc(obj.Spec.RateLimit.QPS, func(v int) types.Int64 { return types.Int64Value(int64(v)) }, types.Int64Null),
		RateLimitBurst:        conv.OptionalFunc(obj.Spec.RateLimit.Burst, func(v int) types.Int64 { return types.Int64Value(int64(v)) }, types.Int64Null),
		AllocationURL:         conv.OptionalFunc(obj.Status.Allocation.URL, types.StringValue, types.StringNull),
		AllocationActiveToken: activeToken(obj.Status.Allocation.Tokens),
		AllocationTokens:      conv.EmptyIfNil(conv.ForEachSliceItem(obj.Status.Allocation.Tokens, types.StringValue)),
		RegistryURL:           conv.OptionalFunc(obj.Status.Registration.URL, types.StringValue, types.StringNull),
		RegistryActiveToken:   activeToken(obj.Status.Registration.Tokens),
		RegistryTokens:        conv.EmptyIfNil(conv.ForEachSliceItem(obj.Status.Registration.Tokens, types.StringValue)),
	}
}

// activeToken returns the last token in the list, or null if the list is empty.
func activeToken(tokens []string) types.String {
	if len(tokens) == 0 {
		return types.StringNull()
	}
	return types.StringValue(tokens[len(tokens)-1])
}
