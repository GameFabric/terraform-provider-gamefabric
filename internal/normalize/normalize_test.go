package normalize_test

import (
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/normalize"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

func TestModelWithIgnorePath(t *testing.T) {
	type testDynamicBuffer struct {
		MaxBufferUtilization      types.Int32 `tfsdk:"max_buffer_utilization"`
		DynamicMaxBufferThreshold types.Int32 `tfsdk:"dynamic_max_buffer_threshold"`
		DynamicMinBufferThreshold types.Int32 `tfsdk:"dynamic_min_buffer_threshold"`
	}
	type testObject struct {
		Foo           types.String       `tfsdk:"foo"`
		DynamicBuffer *testDynamicBuffer `tfsdk:"dynamic_buffer"`
		Bar           types.String       `tfsdk:"bar"`
	}

	state := tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"foo": schema.StringAttribute{},
				"dynamic_buffer": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"max_buffer_utilization": schema.Int32Attribute{
							Required: true,
						},
						"dynamic_max_buffer_threshold": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								planmodifiers.NewDynamicMaxBufferThreshold(),
							},
						},
						"dynamic_min_buffer_threshold": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								planmodifiers.NewDynamicMinBufferThreshold(),
							},
						},
					},
				},
				"bar": schema.StringAttribute{},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"dynamic_buffer": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"max_buffer_utilization":       tftypes.Number,
							"dynamic_max_buffer_threshold": tftypes.Number,
							"dynamic_min_buffer_threshold": tftypes.Number,
						},
					},
					"bar": tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "foo"),
				"dynamic_buffer": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"max_buffer_utilization":       tftypes.Number,
							"dynamic_max_buffer_threshold": tftypes.Number,
							"dynamic_min_buffer_threshold": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"max_buffer_utilization":       tftypes.NewValue(tftypes.Number, 10),
						"dynamic_max_buffer_threshold": tftypes.NewValue(tftypes.Number, 20),
						"dynamic_min_buffer_threshold": tftypes.NewValue(tftypes.Number, 30),
					},
				),
				"bar": tftypes.NewValue(tftypes.String, "bar"),
			},
		),
	}

	apiResponse := testObject{
		Foo: types.StringValue("foo"),
		Bar: types.StringValue("bar"),
	}

	diags := normalize.Model(t.Context(), &apiResponse, state)

	require.Empty(t, diags)

	assert.Equal(t, testObject{
		Foo: types.StringValue("foo"),
		Bar: types.StringValue("bar"),
	}, apiResponse)
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
