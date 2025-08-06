package context

import (
	"fmt"

	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/tfutils/conv"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Context struct {
	ClientSet clientset.Interface
}

func NewContext(clientSet clientset.Interface) *Context {
	c := conv.New()
	c.Register(resource.Quantity{}, quantityToModel, quantityFromModel)
	c.Register(intstr.IntOrString{}, intOrStrToModel, intOrStrFromModel)

	return &Context{
		ClientSet: clientSet,
	}
}

func quantityToModel(q any) (attr.Value, error) {
	if q == nil {
		return types.SetNull(types.StringType), nil
	}
	v, ok := q.(resource.Quantity)
	if !ok {
		return nil, fmt.Errorf("expected resource.Quantity, got %T", q)
	}
	return types.StringValue(v.String()), nil
}

func quantityFromModel(v attr.Value) (any, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	strVal, ok := v.(types.String)
	if !ok {
		return nil, fmt.Errorf("expected types.StringType, got %s", v.String())
	}
	q, err := resource.ParseQuantity(strVal.ValueString())
	if err != nil {
		return nil, fmt.Errorf("parsing quantity: %w", err)
	}
	return q, nil
}

func intOrStrToModel(i any) (attr.Value, error) {
	if i == nil {
		return types.SetNull(types.StringType), nil
	}
	v, ok := i.(intstr.IntOrString)
	if !ok {
		return nil, fmt.Errorf("expected intstr.IntOrString, got %T", i)
	}
	return types.StringValue(v.String()), nil
}

func intOrStrFromModel(v attr.Value) (any, error) {
	if v.IsNull() || v.IsUnknown() {
		return intstr.IntOrString{}, nil
	}
	strVal, ok := v.(types.String)
	if !ok {
		return nil, fmt.Errorf("expected types.StringType, got %s", v.String())
	}
	return intstr.Parse(strVal.ValueString()), nil
}
