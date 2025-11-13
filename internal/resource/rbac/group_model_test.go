package rbac

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewGroupModel(t *testing.T) {
	model := newGroupModel(testGroupObject)
	assert.Equal(t, testGroupModel, model)
}

func TestGroupModel_ToObject(t *testing.T) {
	obj := testGroupModel.ToObject()
	assert.Equal(t, testGroupObject, obj)
}

var (
	testGroupObject = &rbacv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-group",
			Labels: map[string]string{
				"label-key": "label-value",
			},
			Annotations: map[string]string{
				"annotation-key": "annotation-value",
			},
		},
		Users: []string{"test1", "test2"},
	}

	testGroupModel = groupModel{
		ID:   types.StringValue("test-group"),
		Name: types.StringValue("test-group"),
		Labels: map[string]types.String{
			"label-key": types.StringValue("label-value"),
		},
		Annotations: map[string]types.String{
			"annotation-key": types.StringValue("annotation-value"),
		},
		Users: []types.String{types.StringValue("test1"), types.StringValue("test2")},
	}
)
