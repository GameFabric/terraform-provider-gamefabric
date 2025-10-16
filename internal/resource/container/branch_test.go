package container_test

import (
	"fmt"
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBranch(t *testing.T) {
	name := "dflt"
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
					resource.TestCheckResourceAttr("gamefabric_branch.test", "retention_policy_rules.#", "0"),
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
