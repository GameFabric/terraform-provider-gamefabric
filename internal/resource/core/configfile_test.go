package core_test

import (
	"fmt"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestConfigFile(t *testing.T) {
	name := "test-config-file"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceConfigFileDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigFileConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "data", "config file data"),
				),
			},
			{
				ResourceName:      "gamefabric_configfile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceConfigFileConfigBasicWithDescription(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "description", "My Config File Description"),
					resource.TestCheckResourceAttr("gamefabric_configfile.test", "data", "config file data"),
				),
			},
		},
	})
}

func testResourceConfigFileConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_configfile" "test" {
  name = "%s"
  environment = "dflt"
  data = "config file data"
}`, name)
}

func testResourceConfigFileConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_configfile" "test" {
  name = "%s"
  environment = "dflt"
  description = "My Config File Description"
  data = "config file data"
}`, name)
}

func testResourceConfigFileDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_configfile" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.CoreV1().ConfigFiles(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil {
				if resp.Name == name {
					return fmt.Errorf("config file still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
