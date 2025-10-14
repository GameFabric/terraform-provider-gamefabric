package core_test

import (
	"fmt"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestEnvironment(t *testing.T) {
	t.Parallel()

	name := "dflt"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceEnvironmentDestroy(t, cs),
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
  labels = {
	"env" = "test"
  }
}`, name)
}

func testResourceEnvironmentConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_environment" "test" {
  name = "%s"
  display_name = "My Env"
  description = "My Env Description"
}`, name)
}

func testResourceEnvironmentDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_environment" {
				continue
			}

			resp, err := cs.CoreV1().Environments().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil {
				if resp.Name == rs.Primary.ID {
					return fmt.Errorf("environment still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
