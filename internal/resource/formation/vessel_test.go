package formation_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/formation"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators/validatorstest"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestResourceVessel(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckVesselDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceVesselConfigFull(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "name", "my-vessel"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "description", "My Vessel"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "suspend", "true"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "labels.example1", "vessel-label"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "annotations.example2", "vessel-annotation"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "region", "eu"),

					// Game server labels & annotations.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "gameserver_labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "gameserver_labels.example3", "gameserver-label"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "gameserver_annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "gameserver_annotations.example4", "gameserver-annotation"),

					// Containers.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.name", "example-container"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.image_ref.name", "gameserver-asoda0s"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.image_ref.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.command.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.command.0", "example-command"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.args.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.args.0", "example-arg"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.resources.limits.cpu", "500m"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.resources.limits.memory", "512Mi"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.resources.requests.cpu", "250m"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.resources.requests.memory", "256Mi"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.#", "3"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.0.name", "EXAMPLE_ENV"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.0.value", "example_value"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.1.name", "EXAMPLE_CONFIG_FILE"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.1.value_from.config_file", "example_config_file"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.2.name", "EXAMPLE_POD_FIELD"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.envs.2.value_from.field_path", "metadata.name"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.ports.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.ports.0.name", "http"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.ports.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.ports.0.container_port", "8080"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.ports.0.policy", "Passthrough"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.ports.0.protection_protocol", "something"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.volume_mounts.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.volume_mounts.0.name", "example-volume"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.volume_mounts.0.mount_path", "/data/empty"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.volume_mounts.1.name", "persistent-volume"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.volume_mounts.1.mount_path", "/data/pers"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.config_files.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.config_files.0.name", "example-config-file"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.config_files.0.mount_path", "/config/example-config-file"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.secrets.0.name", "example-secret-1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.secrets.0.mount_path", "/secrets/example-secret-1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.secrets.1.name", "example-secret-2"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.secrets.1.mount_path", "/secrets/example-secret-2"),

					// Health checks.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "health_checks.disabled", "false"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "health_checks.initial_delay_seconds", "15"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "health_checks.period_seconds", "10"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "health_checks.failure_threshold", "5"),

					// Termination configuration.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "termination_configuration.grace_period_seconds", "30"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "termination_configuration.maintenance_seconds", "3600"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "termination_configuration.spec_change_seconds", "120"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "termination_configuration.user_initiated_seconds", "60"),

					// Volumes.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "volumes.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "volumes.0.name", "example-volume"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "volumes.0.empty_dir.size_limit", "1Gi"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "volumes.1.name", "persistent-volume"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "volumes.1.persistent.volume_name", "perm-vol"),

					// Gateway policies.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "gateway_policies.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "gateway_policies.0", "test-policy"),

					// Profiling enabled.
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "profiling_enabled", "true"),
				),
			},
			{
				ResourceName: "gamefabric_vessel.test",
				ImportState:  true,
			},
			{
				Config: testResourceVesselConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "name", "my-vessel"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "description", "My New Vessel Description"),
				),
			},
		},
	})
}

func TestResourceVesselConfigBasic(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckVesselDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceVesselConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "name", "my-vessel"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "description", "My New Vessel Description"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "region", "eu"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.name", "example-container"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.image_ref.name", "gameserver-asoda0s"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.image_ref.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.resources.requests.cpu", "250m"),
					resource.TestCheckResourceAttr("gamefabric_vessel.test", "containers.0.resources.requests.memory", "256Mi"),
				),
			},
		},
	})
}

func TestResourceVessel_Validates(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name:        "requires name",
			config:      testResourceVesselConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "name" is required`)),
		},
		{
			name:        "requires environment",
			config:      testResourceVesselConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "environment" is required`)),
		},
		{
			name:        "requires containers",
			config:      testResourceVesselConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "containers" is required`)),
		},
		{
			name:        "requires region",
			config:      testResourceVesselConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "region" is required`)),
		},
		{
			name:        "validates name",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_name!`)),
		},
		{
			name:        "validates empty name",
			config:      testResourceVesselConfigBasicNamed("", "test"),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`name is required`)),
		},
		{
			name:        "validates environment",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`too_long`)),
		},
		{
			name:        "validates region",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_region!`)),
		},
		// Labels.
		{
			name:        "validates label keys",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_label_key!`)),
		},
		{
			name:        "validates label values",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "example1" is not valid`)),
		},
		{
			name:        "validates game server label keys",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_gs_label_key!`)),
		},
		{
			name:        "validates game server label values",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "example3" is not valid`)),
		},
		// Annotations.
		{
			name:        "validates annotation keys",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_annotation_key!`)),
		},
		{
			name:        "validates game server annotation keys",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_gs_annotation_key!`)),
		},
		// Container resources.
		{
			name:        "validates containers[].resources CPU",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`all_cpu`)),
		},
		{
			name:        "validates containers[].resources memory",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`all_memory`)),
		},
		// Container ports.
		{
			name:        "validates containers[].ports protocol",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`INVALID_PROTOCOL`)),
		},
		{
			name:        "validates containers[].ports container port",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`100000`)),
		},
		{
			name:        "validates containers[].ports policy",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_policy`)),
		},
		{
			name:        "validates containers[].ports protection protocol",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_protocol!`)),
		},
		// Health checks.
		{
			name:        "validates health_checks.initial_delay_seconds",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-15`)),
		},
		{
			name:        "validates health_checks.period_seconds",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-14`)),
		},
		{
			name:        "validates health_checks.failure_threshold",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-5`)),
		},
		// Termination grace period.
		{
			name:        "validates termination_configuration.grace_period_seconds",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-30`)),
		},
		{
			name:        "validates termination_configuration.maintenance_seconds",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-3600`)),
		},
		{
			name:        "validates termination_configuration.spec_change_seconds",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-120`)),
		},
		{
			name:        "validates termination_configuration.user_initiated_seconds",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-60`)),
		},
		// Volumes.
		{
			name:        "validates volumes.empty_dir.size_limit",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`around_one_dvd`)),
		},
		{
			name:        "validates volumes.persistent.volume_name",
			config:      testResourceVesselConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_persistent_volume_name!`)),
		},
		// Gateway policies.
		{
			name:        "validates gateway policies",
			config:      testResourceVesselConfigFullInvalid(),
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
				CheckDestroy:             testCheckVesselDestroy(t, cs),
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

func TestVesselResourceGameFabricValidators(t *testing.T) {
	t.Parallel()

	resp := &tfresource.SchemaResponse{}

	arm := formation.NewVessel()
	arm.Schema(t.Context(), tfresource.SchemaRequest{}, resp)

	want := validatorstest.CollectJSONPaths(&formationv1.Vessel{})
	got := validatorstest.CollectPathExpressions(resp.Schema)

	require.NotEmpty(t, got)
	require.NotEmpty(t, want)
	for _, path := range got {
		require.Containsf(t, want, path, "The validator path %q was not found in the Vessel API object", path)
	}
}

func testResourceVesselConfigEmpty() string {
	return `resource "gamefabric_vessel" "test" {}`
}

func testResourceVesselConfigBasicNamed(name, env string, extras ...string) string {
	return fmt.Sprintf(`resource "gamefabric_vessel" "test" {
  name        = %q
  environment = %q
  description = "My New Vessel Description"
  region = "eu"

  containers = [
    {
      name = "example-container"
      image_ref = {
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
}`, name, env, strings.Join(extras, "\n"))
}

func testResourceVesselConfigBasic(extras ...string) string {
	return testResourceVesselConfigBasicNamed("my-vessel", "test", extras...)
}
func testResourceVesselConfigFull() string {
	return `resource "gamefabric_vessel" "test" {
  name        = "my-vessel"
  environment = "test"
  description = "My Vessel"
  suspend     = true

  labels      = {
    "example1": "vessel-label"
  }
  annotations = {
	"example2": "vessel-annotation"
  }
  region = "eu"

  gameserver_labels = {
    "example3": "gameserver-label"
  }
  gameserver_annotations = {
    "example4": "gameserver-annotation"
  }

  containers = [
    {
      name = "example-container"
      image_ref = {
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
          mount_path    = "/data/empty"
        },
        {
          name          = "persistent-volume"
          mount_path    = "/data/pers"
        }
      ]
      config_files = [
        {
          name       = "example-config-file"
          mount_path = "/config/example-config-file"
        }
      ]
      secrets = [
		{
		  name       = "example-secret-1"
		  mount_path = "/secrets/example-secret-1"
		},
		{
		  name       = "example-secret-2"
		  mount_path = "/secrets/example-secret-2"
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
    grace_period_seconds   = 30
    maintenance_seconds    = 3600
    spec_change_seconds    = 120
    user_initiated_seconds = 60
  }

  // Volumes
  volumes = [
    {
      name = "example-volume"
      empty_dir = {
        size_limit = "1Gi"
      }
    },
    {
      name = "persistent-volume"
	  persistent = {
		volume_name = "perm-vol"
	  }
	}
  ]
  gateway_policies  = ["test-policy"]
  profiling_enabled = true
}`
}

func testResourceVesselConfigFullInvalid() string {
	return `resource "gamefabric_vessel" "test" {
  name        = "invalid_name!"
  environment = "too_long"

  labels      = {
    "invalid_label_key!": "vessel-label"
    "example1": "invalid_label_value!"
  }
  annotations = {
    "invalid_annotation_key!": "vessel-annotation"
    "example2": "valid_annotation_value!"
  }
  region = "invalid_region!"

  gameserver_labels = {
    "invalid_gs_label_key!": "gs-label"
    "example3": "invalid_gs_label_value!"
  }
  gameserver_annotations = {
    "invalid_gs_annotation_key!": "gs-annotation"
    "example4": "valid_gs_annotation_value!"
  }

  containers = [
    {
      name = "name"
      image_ref = {
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
    grace_period_seconds   = -30
    maintenance_seconds    = -3600
    spec_change_seconds    = -120
    user_initiated_seconds = -60
  }

  // Volumes
  volumes = [
    {
      name = "example-volume"
      empty_dir = {
        size_limit = "around_one_dvd"
      }
    },
    {
      name = "persistent-volume"
	  persistent = {
		volume_name = "invalid_persistent_volume_name!"
	  }
    }
  ]
  gateway_policies  = ["invalid_gateway_policy!"]
}`
}

func testCheckVesselDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_vessel" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.FormationV1().Vessels(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil && resp.Name == name {
				return fmt.Errorf("vessel still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
