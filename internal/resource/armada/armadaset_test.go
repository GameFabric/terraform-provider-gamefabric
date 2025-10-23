package armada_test

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

func TestResourceArmadaSet(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckArmadaSetDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceArmadaSetConfigFull(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "name", "my-armadaset"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "description", "My ArmadaSet"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "labels.label-key", "armadaset-label"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "annotations.annotation-key", "armadaset-annotation"),

					// Autoscaling.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "autoscaling.fixed_interval_seconds", "10"),

					// Regions.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.name", "eu"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.0.region_type", "baremetal"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.0.min_replicas", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.0.max_replicas", "2"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.0.buffer_size", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.1.region_type", "gcp"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.1.min_replicas", "2"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.1.max_replicas", "5"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.replicas.1.buffer_size", "2"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.envs.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.envs.0.name", "REGION_ENV"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.envs.0.value", "eu_value"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.0.labels.region-label-key", "region-label-value"),

					// Game server labels & annotations.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "gameserver_labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "gameserver_labels.gs-label-key", "gs-label-value"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "gameserver_annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "gameserver_annotations.gs-annotation-key", "gs-annotation-value"),

					// Containers.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.name", "example-container"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.image.name", "gameserver-asoda0s"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.image.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.command.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.command.0", "example-command"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.args.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.args.0", "example-arg"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.resources.limits.cpu", "500m"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.resources.limits.memory", "512Mi"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.resources.requests.cpu", "250m"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.resources.requests.memory", "256Mi"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.#", "3"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.0.name", "EXAMPLE_ENV"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.0.value", "example_value"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.1.name", "EXAMPLE_CONFIG_FILE"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.1.value_from.config_file", "example_config_file"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.2.name", "EXAMPLE_POD_FIELD"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.envs.2.value_from.field_path", "metadata.name"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.ports.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.ports.0.name", "http"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.ports.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.ports.0.container_port", "8080"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.ports.0.policy", "Passthrough"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.ports.0.protection_protocol", "something"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.volume_mounts.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.volume_mounts.0.name", "example-volume"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.volume_mounts.0.mount_path", "/data"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.config_files.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.config_files.0.name", "example-config-file"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.0.config_files.0.mount_path", "/config/example-config-file"),

					// Health checks.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "health_checks.disabled", "false"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "health_checks.initial_delay_seconds", "15"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "health_checks.period_seconds", "10"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "health_checks.failure_threshold", "5"),

					// Termination configuration.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "termination_configuration.grace_period_seconds", "30"),

					// Deployment strategy.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "strategy.rolling_update.max_surge", "25%"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "strategy.rolling_update.max_unavailable", "5"),

					// Volumes.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "volumes.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "volumes.0.name", "example-volume"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "volumes.0.empty_dir.size_limit", "1Gi"),

					// Gateway policies.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "gateway_policies.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "gateway_policies.0", "test-policy"),

					// Profiling enabled.
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "profiling_enabled", "true"),
				),
			},
			{
				ResourceName:      "gamefabric_armadaset.test",
				ImportState:       true,
				ImportStateVerify: true,
			}, {
				Config: testResourceArmadaSetConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "name", "my-armadaset"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "description", "My New ArmadaSet Description"),
				),
			},
		},
	})
}

func TestResourceArmadaConfigSetBasic(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckArmadasDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceArmadaSetConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "name", "my-armadaset"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "description", "My New ArmadaSet Description"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "regions.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "containers.#", "1"),
				),
			},
			{
				ResourceName:      "gamefabric_armadaset.test",
				ImportState:       true,
				ImportStateVerify: true,
			}, {
				Config: testResourceArmadaSetConfigBasic("strategy = {\nrecreate = {}\n}\n"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "name", "my-armadaset"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "description", "My New ArmadaSet Description"),
					resource.TestCheckResourceAttr("gamefabric_armadaset.test", "strategy.recreate.%", "0"),
				),
			},
		},
	})
}

func TestResourceArmadaSet_Validates(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name:        "requires name",
			config:      testResourceArmadaSetConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "name" is required`)),
		},
		{
			name:        "requires environment",
			config:      testResourceArmadaSetConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "environment" is required`)),
		},
		{
			name:        "requires containers",
			config:      testResourceArmadaSetConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "containers" is required`)),
		},
		{
			name:        "validates name",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_name!`)),
		},
		{
			name:        "validates environment",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`too_long`)),
		},
		// Labels.
		{
			name:        "validates label keys",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid-label-key!`)),
		},
		{
			name:        "validates label values",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "valid-label-key" is not valid`)),
		},
		{
			name:        "validates game server label keys",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid-gs-label-key!`)),
		},
		{
			name:        "validates game server label values",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "valid-gs-label-key" is not valid`)),
		},
		{
			name:        "validates region label keys",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid-region-label-key!`)),
		},
		{
			name:        "validates region label values",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "valid-region-label-key" is not valid`)),
		},
		// Annotations.
		{
			name:        "validates annotation keys",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid-annotation-key!`)),
		},
		{
			name:        "validates game server annotation keys",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid-gs-annotation-key!`)),
		},
		// Autoscaling.
		{
			name:        "validates autoscaling.fixed_interval_seconds",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-10`)),
		},
		// Replicas.
		{
			name:        "validates regions[].replicas[].region_type",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_region_type!`)),
		},
		{
			name:        "validates regions[].replicas[].min_replicas",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-11`)),
		},
		{
			name:        "validates regions[].replicas[].max_replicas",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-22`)),
		},
		{
			name:        "validates regions[].replicas[].buffer_size",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-12`)),
		},
		// Container resources.
		{
			name:        "validates containers[].resources CPU",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`all_cpu`)),
		},
		{
			name:        "validates containers[].resources memory",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`all_memory`)),
		},
		// Container ports.
		{
			name:        "validates containers[].ports protocol",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`INVALID_PROTOCOL`)),
		},
		{
			name:        "validates containers[].ports container port",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`100000`)),
		},
		{
			name:        "validates containers[].ports policy",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_policy`)),
		},
		{
			name:        "validates containers[].ports protection protocol",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_protocol!`)),
		},
		// Health checks.
		{
			name:        "validates health_checks.initial_delay_seconds",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-15`)),
		},
		{
			name:        "validates health_checks.period_seconds",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-14`)),
		},
		{
			name:        "validates health_checks.failure_threshold",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-5`)),
		},
		// Termination grace period.
		{
			name:        "validates termination_configuration.grace_period_seconds",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-30`)),
		},
		// Strategy.
		{
			name:        "validates strategy.rolling_update.max_surge",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`200%`)),
		},
		{
			name:        "validates strategy.max_unavailable.max_surge",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-123`)),
		},
		{
			name:        "validates either recreate or rolling_update, but not both",
			config:      testResourceArmadaConfigBasic("strategy = {\nrolling_update = {}\nrecreate = {}\n}\n"),
			expectError: regexp.MustCompile(`(?s)strategy\.recreate.*strategy\.rolling_update`),
		},
		// Volumes.
		{
			name:        "validates volumes.empty_dir.size_limit",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`around_one_dvd`)),
		},
		// Gateway policies.
		{
			name:        "validates gateway policies",
			config:      testResourceArmadaSetConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_gateway_policy!`)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			pf, cs := providertest.ProtoV6ProviderFactories(t)

			resource.Test(t, resource.TestCase{
				IsUnitTest:               true,
				ProtoV6ProviderFactories: pf,
				CheckDestroy:             testCheckArmadasDestroy(t, cs),
				Steps: []resource.TestStep{
					{
						Config:      test.config,
						ExpectError: test.expectError,
					},
				},
			})
		})
	}
}

func testResourceArmadaSetConfigEmpty() string {
	return `resource "gamefabric_armadaset" "test" {}`
}

func testResourceArmadaSetConfigBasic(extras ...string) string {
	return fmt.Sprintf(`resource "gamefabric_armadaset" "test" {
  name        = "my-armadaset"
  environment = "test"
  description = "My New ArmadaSet Description"
  regions = [
    {
      name = "eu"
      replicas = [
        {
          region_type  = "baremetal"
          min_replicas = 1
          max_replicas = 2
          buffer_size  = 1
        }
      ]
    }
  ]
  containers = [
    {
      name = "example-container"
      image = {
        name   = "gameserver-asoda0s"
        branch = "prod"
      }
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
    }
  ]

  %s
}`, strings.Join(extras, "\n"))
}

func testResourceArmadaSetConfigFull() string {
	return `resource "gamefabric_armadaset" "test" {
  name        = "my-armadaset"
  environment = "test"
  description = "My ArmadaSet"

  labels = {
    "label-key" = "armadaset-label"
  }
  annotations = {
	"annotation-key" = "armadaset-annotation"
  }
  autoscaling = {
    fixed_interval_seconds = 10
  }
  regions = [
    {
      name = "eu"
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
      envs = [
        {
          name  = "REGION_ENV"
          value = "eu_value"
        }
      ]
      labels = {
        "region-label-key" = "region-label-value"
      }
    }
  ]

  gameserver_labels = {
    "gs-label-key" = "gs-label-value"
  }
  gameserver_annotations = {
    "gs-annotation-key" = "gs-annotation-value"
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
  health_checks = {
    disabled              = false
    initial_delay_seconds = 15
    period_seconds        = 10
    failure_threshold     = 5
  }

  // Termination Grace Period
  termination_configuration = {
    grace_period_seconds = 30
  }


  // Rollout strategy
  strategy = {
    rolling_update = {
      max_surge       = "25%"
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
}`
}

func testResourceArmadaSetConfigFullInvalid() string {
	return `resource "gamefabric_armadaset" "test" {
  name        = "invalid_name!"
  environment = "too_long"

  labels      = {
    "invalid-label-key!" = "armadaset-label"
    "valid-label-key"    = "invalid-label-value!"
  }
  annotations = {
    "invalid-annotation-key!" = "armada-annotation"
    "valid-annotation-key"    = "valid-annotation-value!"
  }
  autoscaling = {
    fixed_interval_seconds = -10
  }
  regions = [
    {
      name = "invalid_region!"
      replicas = [
        {
          region_type  = "invalid_region_type!"
          min_replicas = -11
          max_replicas = -22
          buffer_size  = -12
        },
      ]
      envs = [
        {
          name  = "INVALID_ENV!"
          value = "valid_value"
        }
      ]
      labels = {
        "invalid-region-label-key!" = "region-label"
        "valid-region-label-key" = "invalid-region-label-value!"
      }
    }
  ]

  gameserver_labels = {
    "invalid-gs-label-key!" = "gs-label"
    "valid-gs-label-key"    = "invalid-gs-label-value!"
  }
  gameserver_annotations = {
    "invalid-gs-annotation-key!" = "gs-annotation"
    "valid-gs-annotation-key"    = "valid-gs-annotation-value!"
  }

  containers = [
    {
      name = "name"
      image = {
        name   = "gameserver-asoda0s"
        branch = "prod"
      }
      command = ["example-command"]
      args    = ["example-arg"]
      resources = {
        limits = {
          cpu    = "all_cpu"
          memory = "all_memory"
        }
      }
      ports = [
        {
          name                = "http"
          protocol            = "INVALID_PROTOCOL"
          container_port      = 100000
          policy              = "invalid_policy"
          protection_protocol = "invalid_protocol!"
        }
      ]
    }
  ]

  // Agones Health Checks
  health_checks = {
    initial_delay_seconds = -15
    period_seconds        = -14
    failure_threshold     = -5
  }

  // Termination Grace Period
  termination_configuration = {
    grace_period_seconds = -30
  }


  // Rollout strategy
  strategy = {
    rolling_update = {
      max_surge       = "200%"
      max_unavailable = "-123"
    }
  }

  // Volumes
  volumes = [
    {
      name = "example-volume"
      empty_dir = {
        size_limit = "around_one_dvd"
      }
    }
  ]
  gateway_policies  = ["invalid_gateway_policy!"]
}`
}

func testCheckArmadaSetDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_armadaset" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.ArmadaV1().ArmadaSets(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil {
				if resp.Name == rs.Primary.ID {
					return fmt.Errorf("armadaset still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
