package core_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestRegion(t *testing.T) {
	t.Parallel()

	name := "eu"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceRegionDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceRegionConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_region.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.name", "baremetal"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.locations.0", "loc-1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.locations.1", "loc-2"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.0.name", "ENV_VAR_1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.0.value", "value1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.1.name", "ENV_VAR_2"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.1.value_from.field_path", "metadata.name"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.2.name", "ENV_VAR_3"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.2.value_from.config_file", "config-file-name"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.scheduling", "Distributed"),
				),
			},
			{
				ResourceName:      "gamefabric_region.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceRegionConfigBasicWithDescription(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_region.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_region.test", "display_name", "My Region"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "description", "My Region Description"),
				),
			},
			{
				Config: testResourceRegionConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.3.value_from.secret.key", "secret"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.0.envs.3.value_from.secret.name", "secret-name"),
				),
			},
		},
	})
}
func TestRegion_Errored(t *testing.T) {
	t.Parallel()

	name := "eu"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceRegionDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      testResourceRegionConfigSecretAndConfigFile(name),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config:      testResourceRegionConfigSecretAndFieldPath(name),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config:      testResourceRegionConfigFileAndFieldPath(name),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

func testResourceRegionConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  environment = "dflt"
  types = [
    {
      name = "baremetal"
      locations = ["loc-1", "loc-2"]
      envs = [{
        name  = "ENV_VAR_1"
        value = "value1"
      }, {
        name  = "ENV_VAR_2"
        value_from = {
          field_path = "metadata.name"
        }
      }, {
        name = "ENV_VAR_3"
        value_from = {
          config_file = "config-file-name"
        }
      }]
      scheduling = "Distributed"
    }
  ]
}`, name)
}

func testResourceRegionConfigFull(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  environment = "dflt"
  types = [
    {
      name = "baremetal"
      locations = ["loc-1", "loc-2"]
      envs = [{
        name  = "ENV_VAR_1"
        value = "value1"
      }, {
        name  = "ENV_VAR_2"
        value_from = {
          field_path = "metadata.name"
        }
      }, {
        name = "ENV_VAR_3"
        value_from = {
          config_file = "config-file-name"
        }
      }, {
        name = "ENV_VAR_4"
        value_from = {
          secret = {
            key = "secret"
            name = "secret-name"
          }
        }
      }]
      scheduling = "Distributed"
    }
  ]
}`, name)
}

func testResourceRegionConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  description = "My Region Description"
  environment = "dflt"
  types = [
    {
      name = "baremetal"
      locations = ["loc-1", "loc-2"]
    }
  ]
}`, name)
}

func testResourceRegionConfigSecretAndConfigFile(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  environment = "dflt"
  types = [
    {
      name = "baremetal"
      locations = ["loc-1", "loc-2"]
      envs = [{
        name  = "ENV_VAR_1"
        value_from = {
          secret = {
			key = "key-name"
			name = "secret-name"
		  }
		  config_file = "config-file-name"
        }
      }]
      scheduling = "Distributed"
    }
  ]
}`, name)
}

func testResourceRegionConfigSecretAndFieldPath(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  environment = "dflt"
  types = [
    {
      name = "baremetal"
      locations = ["loc-1", "loc-2"]
      envs = [{
        name  = "ENV_VAR_1"
        value_from = {
          secret = {
			key = "key-name"
			name = "secret-name"
		  }
		  field_path = "field-path-name"
        }
      }]
      scheduling = "Distributed"
    }
  ]
}`, name)
}

func testResourceRegionConfigFileAndFieldPath(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  environment = "dflt"
  types = [
    {
      name = "baremetal"
      locations = ["loc-1", "loc-2"]
      envs = [{
        name  = "ENV_VAR_1"
        value_from = {
          field_path = "field-path-name"
		  config_file = "config-file-name"
        }
      }]
      scheduling = "Distributed"
    }
  ]
}`, name)
}

func testResourceRegionDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_region" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.CoreV1().Regions(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil && resp.Name == name {
				return fmt.Errorf("region still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
