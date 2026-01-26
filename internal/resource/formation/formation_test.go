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

func TestResourceFormation(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckFormationDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceFormationConfigFull(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_formation.test", "name", "my-formation"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "description", "My Formation Description"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "labels.example1", "formation-label"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "annotations.example2", "formation-annotation"),

					// Game server labels & annotations.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "gameserver_labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "gameserver_labels.example3", "gameserver-label"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "gameserver_annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "gameserver_annotations.example4", "gameserver-annotation"),

					// Volume templates.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volume_templates.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volume_templates.0.name", "perm-vol"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volume_templates.0.reclaim_policy", "Retain"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volume_templates.0.volume_store_name", "example-volume-store"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volume_templates.0.capacity", "5Gi"),

					// Vessels.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.name", "default-vessel"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.region", "eu"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.description", "Default vessel description"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.suspend", "false"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.gameserver_labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.gameserver_labels.vessel-override-label", "override-value"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.command.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.command.0", "start.sh"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.args.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.args.0", "--enable-feature"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.args.1", "--count=10"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.envs.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.envs.0.name", "OVERRIDE_ENV"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.override.containers.0.envs.0.value", "override_value"),

					// Containers.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.name", "example-container"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.image_ref.name", "gameserver-asoda0s"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.image_ref.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.command.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.command.0", "example-command"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.args.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.args.0", "example-arg"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.resources.limits.cpu", "1000m"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.resources.limits.memory", "512Mi"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.resources.requests.cpu", "250m"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.resources.requests.memory", "256Mi"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.#", "3"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.0.name", "EXAMPLE_ENV"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.0.value", "example_value"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.1.name", "EXAMPLE_CONFIG_FILE"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.1.value_from.config_file", "example_config_file"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.2.name", "EXAMPLE_POD_FIELD"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.envs.2.value_from.field_path", "metadata.name"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.ports.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.ports.0.name", "http"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.ports.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.ports.0.container_port", "8080"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.ports.0.policy", "Passthrough"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.ports.0.protection_protocol", "something"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.volume_mounts.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.volume_mounts.0.name", "example-volume"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.volume_mounts.0.mount_path", "/data/temp"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.volume_mounts.1.name", "persistent-volume"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.volume_mounts.1.mount_path", "/data/save_files"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.config_files.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.config_files.0.name", "example-config-file"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.config_files.0.mount_path", "/config/example-config-file"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.secrets.0.name", "mount-1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.secrets.0.mount_path", "/config/example-config-file-1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.secrets.1.name", "mount-2"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.secrets.1.mount_path", "/config/example-config-file-2"),

					// Health checks.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "health_checks.disabled", "false"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "health_checks.initial_delay_seconds", "15"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "health_checks.period_seconds", "10"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "health_checks.failure_threshold", "5"),

					// Termination configuration.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "termination_configuration.grace_period_seconds", "30"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "termination_configuration.maintenance_seconds", "3600"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "termination_configuration.spec_change_seconds", "120"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "termination_configuration.user_initiated_seconds", "60"),

					// Volumes.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volumes.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volumes.0.name", "example-volume"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volumes.0.empty_dir.size_limit", "1Gi"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volumes.1.name", "persistent-volume"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "volumes.1.persistent.volume_name", "perm-vol"),

					// Gateway policies.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "gateway_policies.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "gateway_policies.0", "test-policy"),

					// Profiling enabled.
					resource.TestCheckResourceAttr("gamefabric_formation.test", "profiling_enabled", "true"),
				),
			},
			{
				ResourceName: "gamefabric_formation.test",
				ImportState:  true,
			}, {
				Config: testResourceFormationConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_formation.test", "name", "my-formation"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "description", "My New Formation Description"),
				),
			},
		},
	})
}

func TestResourceFormationConfigBasic(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckFormationDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceFormationConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_formation.test", "name", "my-formation"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "environment", "test"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "description", "My New Formation Description"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.name", "default-vessel"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "vessels.0.region", "eu"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.name", "example-container"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.image_ref.name", "gameserver-asoda0s"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.image_ref.branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.resources.requests.cpu", "250m"),
					resource.TestCheckResourceAttr("gamefabric_formation.test", "containers.0.resources.requests.memory", "256Mi"),
				),
			},
		},
	})
}

func TestResourceFormation_Validates(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name:        "requires name",
			config:      testResourceFormationConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "name" is required`)),
		},
		{
			name:        "requires environment",
			config:      testResourceFormationConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "environment" is required`)),
		},
		{
			name:        "requires containers",
			config:      testResourceFormationConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "containers" is required`)),
		},
		{
			name:        "requires vessels",
			config:      testResourceFormationConfigEmpty(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`The argument "vessels" is required`)),
		},
		{
			name:        "validates name",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_name!`)),
		},
		{
			name:        "validates environment",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`too_long`)),
		},
		// Labels.
		{
			name:        "validates label keys",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_label_key!`)),
		},
		{
			name:        "validates label values",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "example1" is not valid`)),
		},
		{
			name:        "validates game server label keys",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_gs_label_key!`)),
		},
		{
			name:        "validates game server label values",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "example3" is not valid`)),
		},
		{
			name:        "validates override label keys",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_override_label_key!`)),
		},
		{
			name:        "validates override label values",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`Label value for key "example5" is not valid`)),
		},
		// Annotations.
		{
			name:        "validates annotation keys",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_annotation_key!`)),
		},
		{
			name:        "validates game server annotation keys",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_gs_annotation_key!`)),
		},
		// Volume templates.
		{
			name:        "validates volume template name",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_volume_template_name!`)),
		},
		{
			name:        "validates volume template reclaim policy",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`InvalidReclaimPolicy`)),
		},
		{
			name:        "validates volume template volume store name",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_volume_store_name!`)),
		},
		{
			name:        "validates volume template capacity",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`everything`)),
		},
		// Vessels.
		{
			name:        "validates vessel name",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_vessel_name!`)),
		},
		{
			name:        "validates vessel region",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_region!`)),
		},

		// Container resources.
		{
			name:        "validates containers[].resources CPU",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`all_cpu`)),
		},
		{
			name:        "validates containers[].resources memory",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`all_memory`)),
		},
		// Container ports.
		{
			name:        "validates containers[].ports protocol",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`INVALID_PROTOCOL`)),
		},
		{
			name:        "validates containers[].ports container port",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`100000`)),
		},
		{
			name:        "validates containers[].ports policy",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_policy`)),
		},
		{
			name:        "validates containers[].ports protection protocol",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_protocol!`)),
		},
		// Health checks.
		{
			name:        "validates health_checks.initial_delay_seconds",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-15`)),
		},
		{
			name:        "validates health_checks.period_seconds",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-14`)),
		},
		{
			name:        "validates health_checks.failure_threshold",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-5`)),
		},
		// Termination grace period.
		{
			name:        "validates termination_configuration.grace_period_seconds",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-30`)),
		},
		{
			name:        "validates termination_configuration.maintenance_seconds",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-3600`)),
		},
		{
			name:        "validates termination_configuration.spec_change_seconds",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-120`)),
		},
		{
			name:        "validates termination_configuration.user_initiated_seconds",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`-60`)),
		},
		// Volumes.
		{
			name:        "validates volumes.empty_dir.size_limit",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`around_one_dvd`)),
		},
		{
			name:        "validates volumes.persistent.volume_name",
			config:      testResourceFormationConfigFullInvalid(),
			expectError: regexp.MustCompile(regexp.QuoteMeta(`invalid_persistent_volume_name!`)),
		},
		// Gateway policies.
		{
			name:        "validates gateway policies",
			config:      testResourceFormationConfigFullInvalid(),
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
				CheckDestroy:             testCheckFormationDestroy(t, cs),
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

func TestFormationResourceGameFabricValidators(t *testing.T) {
	t.Parallel()

	resp := &tfresource.SchemaResponse{}

	arm := formation.NewFormation()
	arm.Schema(t.Context(), tfresource.SchemaRequest{}, resp)

	want := validatorstest.CollectJSONPaths(&formationv1.Formation{})
	got := validatorstest.CollectPathExpressions(resp.Schema)

	require.NotEmpty(t, got)
	require.NotEmpty(t, want)
	for _, path := range got {
		require.Containsf(t, want, path, "The validator path %q was not found in the Formation API object", path)
	}
}

func testResourceFormationConfigEmpty() string {
	return `resource "gamefabric_formation" "test" {}`
}

func testResourceFormationConfigBasic(extras ...string) string {
	return fmt.Sprintf(`resource "gamefabric_formation" "test" {
  name        = "my-formation"
  environment = "test"
  description = "My New Formation Description"
  vessels = [
    {
      name = "default-vessel"
      region = "eu"
    }
  ]

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
}`, strings.Join(extras, "\n"))
}

func testResourceFormationConfigFull() string {
	return `resource "gamefabric_formation" "test" {
  name        = "my-formation"
  environment = "test"
  description = "My Formation Description"

  labels      = {
    "example1": "formation-label"
  }
  annotations = {
	"example2": "formation-annotation"
  }

  gameserver_labels = {
    "example3": "gameserver-label"
  }
  gameserver_annotations = {
    "example4": "gameserver-annotation"
  }

  volume_templates = [
    {
      name = "perm-vol"
      reclaim_policy = "Retain"
      volume_store_name = "example-volume-store"
      capacity = "5Gi"
    }
  ]

  vessels = [
    {
      name = "default-vessel"
      region = "eu"
      description = "Default vessel description"
      suspend = false
      override = { 
        gameserver_labels = {
          "vessel-override-label": "override-value"
        }
        containers = [{
          command = ["start.sh"]
          args = ["--enable-feature", "--count=10"]
          envs = [ 
            {
              name  = "OVERRIDE_ENV"
              value = "override_value"
            }
          ]
        }]
      }
    }
  ]

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
          cpu    = "1000m"
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
          mount_path    = "/data/temp"
        },
        {
          name          = "persistent-volume"
          mount_path    = "/data/save_files"
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
          name       = "mount-1"
          mount_path = "/config/example-config-file-1"
        },
        {
          name       = "mount-2"
          mount_path = "/config/example-config-file-2"
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

func testResourceFormationConfigFullInvalid() string {
	return `resource "gamefabric_formation" "test" {
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

  gameserver_labels = {
    "invalid_gs_label_key!": "gs-label"
    "example3": "invalid_gs_label_value!"
  }
  gameserver_annotations = {
    "invalid_gs_annotation_key!": "gs-annotation"
    "example4": "valid_gs_annotation_value!"
  }

  volume_templates = [
    {
      name = "invalid_volume_template_name!"
      reclaim_policy = "InvalidReclaimPolicy"
      volume_store_name = "invalid_volume_store_name!"
      capacity = "everything"
    },
  ]

  vessels = [
    {
      name = "invalid_vessel_name!"
      region = "invalid_region!"
      override = { 
        gameserver_labels = {
          "invalid_override_label_key!": "vessel-label"
          "example5": "invalid_override_label_value!"
        }
      }
    }
  ]

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

func testCheckFormationDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_formation" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.FormationV1().Formations(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil && resp.Name == name {
				return fmt.Errorf("formation still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
