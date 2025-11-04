package validatorstest_test

import (
	"context"
	"regexp"
	"testing"

	v1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators/validatorstest"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestCollectCollectJSONPaths(t *testing.T) {
	t.Parallel()

	want := []string{
		"apiVersion",
		"data",
		"description",
		"kind",
		"metadata",
		"metadata.annotations",
		"metadata.createdTimestamp",
		"metadata.deletedTimestamp",
		"metadata.environment",
		"metadata.finalizers",
		"metadata.finalizers[?]",
		"metadata.labels",
		"metadata.name",
		"metadata.ownerReferences",
		"metadata.ownerReferences[?]",
		"metadata.ownerReferences[?].apiVersion",
		"metadata.ownerReferences[?].kind",
		"metadata.ownerReferences[?].name",
		"metadata.ownerReferences[?].uid",
		"metadata.revision",
		"metadata.uid",
		"metadata.updatedTimestamp",
		"status",
		"status.lastDataChange",
		"status.state",
	}
	got := validatorstest.CollectJSONPaths(&v1.ConfigFile{})

	assert.Equal(t, want, got)
}

func TestCollectPathExpressions(t *testing.T) {
	t.Parallel()

	want := []string{
		"spec.bool",
		"spec.int32",
		"spec.int64",
		"spec.list",
		"spec.listNested",
		"spec.listNested.int64",
		"spec.map",
		"spec.singleNested",
		"spec.singleNested.string",
		"spec.string",
	}
	got := validatorstest.CollectPathExpressions(schema.Schema{
		Attributes: map[string]schema.Attribute{
			"string": schema.StringAttribute{
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("irrelevant"), "just another validator"),
					validators.GFFieldString(&testValidator{}, "spec.string"),
				},
			},
			"map": schema.MapAttribute{
				Validators: []validator.Map{
					validators.GFFieldMap(&testValidator{}, "spec.map"),
				},
				ElementType: types.StringType,
			},
			"list": schema.ListAttribute{
				Validators: []validator.List{
					validators.GFFieldList(&testValidator{}, "spec.list"),
				},
				ElementType: types.StringType,
			},
			"listNested": schema.ListNestedAttribute{
				Validators: []validator.List{
					validators.GFFieldList(&testValidator{}, "spec.listNested"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"int64": schema.Int64Attribute{
							Validators: []validator.Int64{
								validators.GFFieldInt64(&testValidator{}, "spec.listNested.int64"),
							},
						},
					},
				},
			},
			"singleNested": schema.SingleNestedAttribute{
				Validators: []validator.Object{
					validators.GFFieldObject(&testValidator{}, "spec.singleNested"),
				},
				Attributes: map[string]schema.Attribute{
					"string": schema.StringAttribute{
						Validators: []validator.String{
							validators.GFFieldString(&testValidator{}, "spec.singleNested.string"),
						},
					},
				},
			},
			"bool": schema.BoolAttribute{
				Validators: []validator.Bool{
					validators.GFFieldBool(&testValidator{}, "spec.bool"),
				},
			},
			"int32": schema.Int32Attribute{
				Validators: []validator.Int32{
					validators.GFFieldInt32(&testValidator{}, "spec.int32"),
				},
			},
			"int64": schema.Int64Attribute{
				Validators: []validator.Int64{
					validators.GFFieldInt64(&testValidator{}, "spec.int64"),
				},
			},
		},
	})

	assert.Equal(t, want, got)
}

type testValidator struct{}

func (v *testValidator) Validate(context.Context, validators.GameFabricValidatorRequest) diag.Diagnostics {
	return nil
}
