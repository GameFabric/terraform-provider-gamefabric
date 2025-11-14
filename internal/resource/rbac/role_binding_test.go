package rbac_test

import (
	"fmt"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestRoleBindingResource(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testRoleBindingResourceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_role_binding" "test" {
  role   = "example-role"
  groups = [""]
  users  = [""]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_role_binding.test", "role", "example-role"),
				),
			},
			{
				Config: `resource "gamefabric_role_binding" "test" {
  role   = "example-role"
  groups = ["group1", "group2"]
  users  = ["user1@example.com"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_role_binding.test", "role", "example-role"),
					resource.TestCheckResourceAttr("gamefabric_role_binding.test", "groups.0", "group1"),
					resource.TestCheckResourceAttr("gamefabric_role_binding.test", "groups.1", "group2"),
					resource.TestCheckResourceAttr("gamefabric_role_binding.test", "users.0", "user1@example.com"),
				),
			},
		},
	})
}

func testRoleBindingResourceDestroy(t *testing.T, cs clientset.Interface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_role_binding" {
				continue
			}

			resp, err := cs.RBACV1().RoleBindings().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil {
				return fmt.Errorf("role binding %q still exists: %+v", rs.Primary.ID, resp)
			}
		}
		return nil
	}
}
