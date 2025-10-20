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
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "labels.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "labels.example", "armada-label"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "annotations.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armada_v1.test", "annotations.example", "armada-annotation"),
				),
			},
			{
				ResourceName:      "gamefabric_region.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testResourceArmadasConfigBasic(env, name string) string {
	return fmt.Sprintf(`resource "gamefabric_armada" "test" {
  name        = %q
  environment = %q
  description = "My Armada"
  labels      = {
    "example": "armada-label"
  }
  annotations = {
	"example": "armada-annotation"
  }
  autoscaling {
    fixed_interval_seconds = 10
  }
  region = "eu"
  replicas = [
    {
      region_type  = "baremetal"
      min_replicas = 1
      max_replicas = 2
      buffer_size  = 1
    },
    {
      region_type  = "gcp"
      min_replicas = 2
      max_replicas = 5
      buffer_size  = 2
    }
  ]

  gameserver_labels      = {
    "example": "gameserver-label"
  }
  gameserver_annotations = {
    "example": "gameserver-annotation"
  }

  containers = [
    {
      name = "example-container"
      image = {
        name   = "gameserver-asoda0s"
        branch = "prod"
      }
      command = ["example-command"]
      args    = ["example-arg"]
      resources = {
        limits = {
          cpu    = "500m"
          memory = "512Mi"
        }
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      envs = [
        {
          name  = "EXAMPLE_ENV"
          value = "example_value"
        },
        {
          name = "EXAMPLE_CONFIG_FILE"
          value_from = {
            config_file = "example_config_file"
          }
        },
        {
          name = "EXAMPLE_POD_FIELD"
          value_from = {
            field_path = "metadata.name"
          }
        }
      ]
      ports = [
        {
          name                = "http"
          protocol            = "TCP"
          container_port      = 8080
          policy              = "Passthrough"
          protection_protocol = "something"
        }
      ]
      volume_mounts = [
        {
          name          = "example-volume"
          mount_path    = "/data"
        }
      ]
      config_files = [
        {
          name       = "example-config-file"
          mount_path = "/config/example-config-file"
        }
      ]
    }
  ]

  // Agones Health Checks
  health_checks {
    disabled              = false
    initial_delay_seconds = 15
    period_seconds        = 10
    failure_threshold     = 5
  }

  // Termination Grace Period
  termination_configuration {
    grace_period_seconds = 30
  }


  // Rollout strategy
  strategy {
    rolling_update {
      max_surge       = "25%%"
      max_unavailable = "5"
    }
  }

  // Volumes
  volumes = [
    {
      name = "example-volume"
      empty_dir = {
        size_limit = "1Gi"
      }
    }
  ]
  gateway_policies  = ["test-policy"]
  profiling_enabled = true

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
