package rbac_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGroup(t *testing.T) {
	t.Parallel()

	grp := &rbacv1.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group"},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, grp)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_group" "test1" {
  name = "test-group"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_group.test1", "name", "test-group"),
				),
			},
		},
	})
}
