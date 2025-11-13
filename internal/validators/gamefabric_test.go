package validators_test

import (
	"errors"
	"testing"

	"github.com/gamefabric/gf-apicore/api/validation"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apicore/runtime"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGamefabricStoreValidator_Validate(t *testing.T) {
	storeVal := fakeStoreValidator{errs: validation.Errors{
		errors.New("field spec.b[0] should be better than spec.c"),
		errors.New("field spec.b[0] should be more different than spec.a"),
		errors.New("field not-spec.b[0] should be ignored"),
		errors.New("field spec.b[0]-not should be ignored"),
	}}
	val := validators.NewGameFabricValidator[*TestObject, TestModel](func() validators.StoreValidator { return storeVal })

	req := validators.GameFabricValidatorRequest{
		ConfigValue: types.Int32Value(42),
		Path:        path.Root("b").AtListIndex(0),
		Config: tfsdk.Config{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"a": schema.StringAttribute{Validators: []validator.String{
						validators.GFFieldString(val, "spec.a"),
					}},
					"b": schema.ListAttribute{
						ElementType: types.Int32Type,
						Validators: []validator.List{
							validators.GFFieldList(val, "spec.b[?]"),
						},
					},
					"c": schema.Int64Attribute{Validators: []validator.Int64{
						validators.GFFieldInt64(val, "spec.c"),
					}},
				},
			},
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.List{ElementType: tftypes.Number},
					"c": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "bar value"),
				"b": tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, []tftypes.Value{tftypes.NewValue(tftypes.Number, 42)}),
				"c": tftypes.NewValue(tftypes.Number, 43),
			}),
		},
		PathExpr: "spec.b[?]",
	}

	got := val.Validate(t.Context(), req)

	if got.ErrorsCount() != 2 {
		t.Fatalf("expected 2 error(s), got %d: %v", got.ErrorsCount(), got)
	}
}

type TestObject struct {
	metav1.TypeMeta

	Spec TestObjectSpec `json:"spec"`
}

type TestObjectSpec struct {
	A string `json:"a"`
	B []int  `json:"b"`
	C int    `json:"c"`
}

type TestModel struct {
	A types.String  `tfsdk:"a"`
	B []types.Int32 `tfsdk:"b"`
	C types.Int64   `tfsdk:"c"`
}

func (m TestModel) ToObject() *TestObject {
	return &TestObject{
		Spec: TestObjectSpec{
			A: m.A.ValueString(),
			B: conv.ForEachSliceItem(m.B, func(i types.Int32) int { return int(i.ValueInt32()) }),
			C: int(m.C.ValueInt64()),
		},
	}
}

type fakeStoreValidator struct {
	errs validation.Errors
}

func (f fakeStoreValidator) Validate(runtime.Object) validation.Errors {
	return f.errs
}
