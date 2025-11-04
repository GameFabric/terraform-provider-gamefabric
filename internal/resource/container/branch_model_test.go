package container

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBranchModel(t *testing.T) {
	model := newBranchModel(testBranchObject)
	assert.Equal(t, testBranchModel, model)
}

func TestBranchModel_ToObject(t *testing.T) {
	obj := testBranchModel.ToObject()
	assert.Equal(t, testBranchObject, obj)
}

var (
	testBranchObject = &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch Display Name",
			Description: "Test Branch Description",
			RetentionPolicyRules: []containerv1.BranchImageRetentionPolicyRule{
				{
					Name:       "rule1",
					ImageRegex: "image-.*",
					TagRegex:   "v[0-9]+.*",
					KeepCount:  10,
					KeepDays:   30,
				},
				{
					Name:       "rule2",
					ImageRegex: ".*",
					TagRegex:   "latest",
					KeepCount:  5,
					KeepDays:   7,
				},
			},
		},
	}

	testBranchModel = branchModel{
		ID:   types.StringValue("test-branch"),
		Name: types.StringValue("test-branch"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		DisplayName: types.StringValue("Test Branch Display Name"),
		Description: types.StringValue("Test Branch Description"),
		RetentionPolicyRules: []branchImageRetentionPolicyRuleModel{
			{
				Name:       types.StringValue("rule1"),
				ImageRegex: types.StringValue("image-.*"),
				TagRegex:   types.StringValue("v[0-9]+.*"),
				KeepCount:  types.Int64Value(10),
				KeepDays:   types.Int64Value(30),
			},
			{
				Name:       types.StringValue("rule2"),
				ImageRegex: types.StringValue(".*"),
				TagRegex:   types.StringValue("latest"),
				KeepCount:  types.Int64Value(5),
				KeepDays:   types.Int64Value(7),
			},
		},
	}
)
