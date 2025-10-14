package core_test

import (
	"fmt"
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRegion(t *testing.T) {
	name := "eu"
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: testResourceRegionConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_region.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.locations.0", "loc-1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.locations.1", "loc-2"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.0.name", "ENV_VAR_1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.0.value", "value1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.1.name", "ENV_VAR_2"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.1.value_from.field_ref.field_path", "metadata.name"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.1.value_from.field_ref.api_version", "v1"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.2.name", "ENV_VAR_3"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.env.2.value_from.config_file_key_ref.name", "config-file-name"),
					resource.TestCheckResourceAttr("gamefabric_region.test", "types.baremetal.scheduling", "Distributed"),
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
		},
	})
}

func testResourceRegionConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  environment = "dflt"
  types = {
    baremetal = {
      locations = ["loc-1", "loc-2"]
      env = [{
        name  = "ENV_VAR_1"
        value = "value1"
      }, {
        name  = "ENV_VAR_2"
        value_from = {
          field_ref = {
            field_path = "metadata.name"
            api_version = "v1"  
          }  
        }
      }, {
        name = "ENV_VAR_3"
        value_from = {
          config_file_key_ref = {
            name = "config-file-name"
          }
        }
      }]
      scheduling = "Distributed"
    }
  }
}`, name)
}

func testResourceRegionConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_region" "test" {
  name = "%s"
  display_name = "My Region"
  description = "My Region Description"
  environment = "dflt"
  types = {
    baremetal = {
      locations = ["loc-1", "loc-2"]
    }
  }
}`, name)
}
