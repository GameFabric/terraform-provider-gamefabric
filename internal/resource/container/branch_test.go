package container_test

import (
	"fmt"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestBranch(t *testing.T) {
	t.Parallel()

	name := "test-branch"
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: testResourceBranchConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_branch.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "display_name", "My Branch"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.#", "0"),
				),
			},
			{
				ResourceName:      "gamefabric_branch.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceBranchConfigBasicWithDescription(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_branch.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "display_name", "My Branch"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "description", "My Branch Description"),
				),
			},
			{
				Config: testResourceBranchConfigWithRetentionPolicy(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_branch.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "display_name", "My Branch"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.0.name", "rule1"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.0.image_regex", ".*"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.0.tag_regex", ".*"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.0.keep_count", "10"),
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.0.keep_days", "30"),
				),
			},
		},
	})
}

func testResourceBranchConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_branch" "test" {
  name = "%s"
  display_name = "My Branch"
  retention_policy_rules = []
}`, name)
}

func testResourceBranchConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_branch" "test" {
  name = "%s"
  display_name = "My Branch"
  description = "My Branch Description"
  retention_policy_rules = []
}`, name)
}

func testResourceBranchConfigWithRetentionPolicy(name string) string {
	return fmt.Sprintf(`resource "gamefabric_branch" "test" {
  name = "%s"
  display_name = "My Branch"
  description = "My Branch Description"
  retention_policy_rules = [
	{
	  name = "rule1"
      keep_count = 10
      keep_days = 30
      image_regex = ".*"
	  tag_regex = ".*"
	}
  ]
}`, name)
}

func testResourceBranchDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_branch" {
				continue
			}

			resp, err := cs.ContainerV1().Branches().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil && resp.Name == rs.Primary.ID {
				return fmt.Errorf("branch still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
