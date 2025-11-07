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

func TestGroup(t *testing.T) {
	t.Parallel()

	name := "test-group"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceGroupDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceGroupConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_group.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_group.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_group.test", "labels.foo", "bar"),
					resource.TestCheckResourceAttr("gamefabric_group.test", "annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_group.test", "annotations.bar", "baz"),
					resource.TestCheckResourceAttr("gamefabric_group.test", "users.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_group.test", "users.0", "user1@example.com"),
					resource.TestCheckResourceAttr("gamefabric_group.test", "users.1", "user2@example.com"),
				),
			},
			{
				ResourceName:      "gamefabric_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceGroupConfigBasicModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_group.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_group.test", "labels.foo", "bar"),
				),
			},
		},
	})
}

func testResourceGroupConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_group" "test" {
  name = "%s"
  labels = {
	foo = "bar"
  }
  annotations = {
	bar = "baz"
  }
  users = ["user1@example.com", "user2@example.com"]
}`, name)
}

func testResourceGroupConfigBasicModified(name string) string {
	return fmt.Sprintf(`resource "gamefabric_group" "test" {
  name = "%s"
  labels = {
	foo = "bar"
  }
}`, name)
}

func testResourceGroupDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_group" {
				continue
			}

			resp, err := cs.RBACV1().Groups().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil {
				if resp.Name == rs.Primary.ID {
					return fmt.Errorf("group still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
