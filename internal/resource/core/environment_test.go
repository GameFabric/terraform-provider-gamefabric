package core_test

import (
	"fmt"
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestEnvironment(t *testing.T) {
	t.Parallel()

	name := "dflt"
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: testResourceEnvironmentConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_environment.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_environment.test", "display_name", "My Env"),
				),
			},
			{
				ResourceName:      "gamefabric_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceEnvironmentConfigBasicWithDescription(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_environment.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_environment.test", "display_name", "My Env"),
					resource.TestCheckResourceAttr("gamefabric_environment.test", "description", "My Env Description"),
				),
			},
		},
	})
}

func testResourceEnvironmentConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_environment" "test" {
  name = "%s"
  display_name = "My Env"
}`, name)
}

func testResourceEnvironmentConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_environment" "test" {
  name = "%s"
  display_name = "My Env"
  description = "My Env Description"
}`, name)
}
