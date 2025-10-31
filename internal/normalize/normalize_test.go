package normalize_test

import (
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelWithEmpty(t *testing.T) {
	state := tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{},
				"sub_model": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"age":  schema.Int64Attribute{},
						"city": schema.StringAttribute{},
					},
				},
				"ptr_sub_model": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"age":  schema.Int64Attribute{},
						"city": schema.StringAttribute{},
					},
				},
				"map_str": schema.MapAttribute{
					ElementType: types.StringType,
				},
				"map_obj": schema.MapNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"age":  schema.Int64Attribute{},
							"city": schema.StringAttribute{},
						},
					},
				},
				"slice_bool": schema.ListAttribute{
					ElementType: types.BoolType,
				},
				"slice_obj": schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"age":  schema.Int64Attribute{},
							"city": schema.StringAttribute{},
						},
					},
				},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
					"sub_model": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					"ptr_sub_model": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					"map_str": tftypes.Map{ElementType: tftypes.String},
					"map_obj": tftypes.Map{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					}},
					"slice_bool": tftypes.List{ElementType: tftypes.Bool},
					"slice_obj": tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					}},
				},
			},
			map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, ""),
				"sub_model": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"age":  tftypes.NewValue(tftypes.Number, 0),
						"city": tftypes.NewValue(tftypes.String, ""),
					},
				),
				"ptr_sub_model": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"age":  tftypes.NewValue(tftypes.Number, 0),
						"city": tftypes.NewValue(tftypes.String, ""),
					},
				),
				"map_str": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{}),
				"map_obj": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"age":  tftypes.Number,
						"city": tftypes.String,
					},
				}}, map[string]tftypes.Value{}),
				"slice_bool": tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{}),
				"slice_obj": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"age":  tftypes.Number,
						"city": tftypes.String,
					},
				}}, []tftypes.Value{}),
			},
		),
	}

	obj := TestObject{
		Name: types.StringNull(),
		SubModel: TestSubModel{
			Age:  types.Int64Null(),
			City: types.StringNull(),
		},
		PtrSubModel: nil,
		MapStr:      nil,
		MapObj:      nil,
		SliceBool:   nil,
		SliceObj:    nil,
	}
	want := TestObject{
		Name: types.StringValue(""),
		SubModel: TestSubModel{
			Age:  types.Int64Value(0),
			City: types.StringValue(""),
		},
		PtrSubModel: &TestSubModel{
			Age:  types.Int64Value(0),
			City: types.StringValue(""),
		},
		MapStr:    map[string]types.String{},
		MapObj:    map[string]TestSubModel{},
		SliceBool: []types.Bool{},
		SliceObj:  []TestSubModel{},
	}

	got := normalize.Model(t.Context(), &obj, state)

	require.Empty(t, got)
	assert.Equal(t, want, obj)
}

func TestModelWithNull(t *testing.T) {
	state := tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{},
				"sub_model": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"age":  schema.Int64Attribute{},
						"city": schema.StringAttribute{},
					},
				},
				"ptr_sub_model": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"age":  schema.Int64Attribute{},
						"city": schema.StringAttribute{},
					},
				},
				"map_str": schema.MapAttribute{
					ElementType: types.StringType,
				},
				"map_obj": schema.MapNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"age":  schema.Int64Attribute{},
							"city": schema.StringAttribute{},
						},
					},
				},
				"slice_bool": schema.ListAttribute{
					ElementType: types.BoolType,
				},
				"slice_obj": schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"age":  schema.Int64Attribute{},
							"city": schema.StringAttribute{},
						},
					},
				},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
					"sub_model": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					"ptr_sub_model": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					"map_str": tftypes.Map{ElementType: tftypes.String},
					"map_obj": tftypes.Map{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					}},
					"slice_bool": tftypes.List{ElementType: tftypes.Bool},
					"slice_obj": tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					}},
				},
			},
			map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, nil),
				"sub_model": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					nil,
				),
				"ptr_sub_model": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"age":  tftypes.Number,
							"city": tftypes.String,
						},
					},
					nil,
				),
				"map_str": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"map_obj": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"age":  tftypes.Number,
						"city": tftypes.String,
					},
				}}, nil),
				"slice_bool": tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, nil),
				"slice_obj": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"age":  tftypes.Number,
						"city": tftypes.String,
					},
				}}, nil),
			},
		),
	}

	obj := TestObject{
		Name: types.StringValue(""),
		SubModel: TestSubModel{
			Age:  types.Int64Value(0),
			City: types.StringValue(""),
		},
		PtrSubModel: &TestSubModel{},
		MapStr:      map[string]types.String{},
		MapObj:      map[string]TestSubModel{},
		SliceBool:   []types.Bool{},
		SliceObj:    []TestSubModel{},
	}
	want := TestObject{
		Name: types.StringNull(),
		SubModel: TestSubModel{
			Age:  types.Int64Null(),
			City: types.StringNull(),
		},
		PtrSubModel: nil,
		MapStr:      nil,
		MapObj:      nil,
		SliceBool:   nil,
		SliceObj:    nil,
	}

	got := normalize.Model(t.Context(), &obj, state)

	require.Empty(t, got)
	assert.Equal(t, want, obj)
}

type TestObject struct {
	Name        types.String            `tfsdk:"name"`
	SubModel    TestSubModel            `tfsdk:"sub_model"`
	PtrSubModel *TestSubModel           `tfsdk:"ptr_sub_model"`
	MapStr      map[string]types.String `tfsdk:"map_str"`
	MapObj      map[string]TestSubModel `tfsdk:"map_obj"`
	SliceBool   []types.Bool            `tfsdk:"slice_bool"`
	SliceObj    []TestSubModel          `tfsdk:"slice_obj"`
}

type TestSubModel struct {
	Age  types.Int64  `tfsdk:"age"`
	City types.String `tfsdk:"city"`
}
