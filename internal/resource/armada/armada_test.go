package armada_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceArmadas(t *testing.T) {
	name := "my-armada"
	env := "dflt"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckArmadasDestroy(cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceArmadasConfigBasic(env, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "environment", env),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "description", "My Armada"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "region", "eu"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.name", "baremetal"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.min_replicas", "1"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.max_replicas", "2"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.buffer_size", "3"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.labels.foo", "bar"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.strategy.rolling_update.max_unavailable", "25%"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.strategy.rolling_update.max_surge", "10"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.name", "my-ctr"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.image", "test-xyz"),
				),
			},
			{
				Config: testResourceArmadasConfigBasicWithEnv(env, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "environment", env),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "description", "My Armada"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "region", "eu"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.name", "baremetal"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.min_replicas", "1"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.max_replicas", "2"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "replicas.0.buffer_size", "3"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.labels.foo", "bar"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.name", "my-ctr"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.image", "test-xyz"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.env.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.env.0.name", "foo"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.env.0.value", "bar"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.env.1.name", "baz"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "template.containers.0.env.1.value_from.config_file_key_ref.name", "bat"),
				),
			},
			// TODO: Add import test when import is supported.
			// {
			// 	ResourceName:      "gamefabric_armada.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func testResourceArmadasConfigBasic(env, name string) string {
	return fmt.Sprintf(`resource "gamefabric_armada_v1" "test" {
  name = "%s"
  environment = "%s"

  description = "My Armada"
  region = "eu"
  replicas = [
    {
      region_type = "baremetal"
      min_replicas = 1
      max_replicas = 2
      buffer_size = 3
    }
  ]

  template = {
    labels = {
      "foo" = "bar"
    }
    strategy = {
      rolling_update = {
        max_unavailable = "25%%"
        max_surge = "10"
      }
    }
    containers = [{
      name = "my-ctr"
      branch = "prod"
      image = "test-xyz"
    }]
  }
}`, name, env)
}

func testResourceArmadasConfigBasicWithEnv(env, name string) string {
	return fmt.Sprintf(`resource "gamefabric_armada_v1" "test" {
  name = "%s"
  environment = "%s"

  description = "My Armada"
  region = "eu"
  replicas = [
    {
      region_type = "baremetal"
      min_replicas = 1
      max_replicas = 2
      buffer_size = 3
    }
  ]

  template = {
    labels = {
      "foo" = "bar"
    }
    containers = [{
      name = "my-ctr"
      branch = "prod"
      image = "test-xyz"
      env = [
        {
          name = "foo"
          value = "bar"
        },
        {
          name = "baz"
          value_from = {
            config_file_key_ref = {
              name = "bat"
            }
          }
        }
      ]
    }]
  }
}`, name, env)
}

func testCheckArmadasDestroy(cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_armada_v1" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.ArmadaV1().Armadas(env).Get(context.Background(), name, metav1.GetOptions{})
			if err == nil {
				if resp.Name == rs.Primary.ID {
					return fmt.Errorf("armada still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
