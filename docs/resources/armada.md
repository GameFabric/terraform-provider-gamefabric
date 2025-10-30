---
page_title: "gamefabric_armada Resource - GameFabric"
subcategory: ""
description: |-
  
---

# gamefabric_armada (Resource)

An Armada manages your game server deployments within a Region. It handles the distribution and scaling of your game servers, whether you're running on a single node or across multiple locations with mixed baremetal and cloud infrastructure.

When you update an Armada, it creates a new revision. This versioning system allows you to safely roll back to a previous configuration if needed, or run multiple versions simultaneously during gradual rollouts to minimize disruption to your players.

For details check the <a href="https://docs.gamefabric.com/multiplayer-servers/getting-started/glossary#armada">GameFabric documentation</a>.


## Example Usage - Simple Armada

This example sets up a minimal armada in a single region for europe which uses baremetal and cloud locations.

```terraform
data "gamefabric_branch" "prod" {
  display_name = "prod"
}

data "gamefabric_image" "gameserver" {
  branch = data.gamefabric_branch.prod.name
  image  = "gameserver"
  tag    = "1.2.3"
}

data "gamefabric_environment" "prod" {
  name = "prod"
}

data "gamefabric_region" "europe" {
  name        = "europe"
  environment = data.gamefabric_environment.prod.name
}

resource "gamefabric_armada" "this" {
  name        = "myarmada"
  environment = data.gamefabric_environment.prod.name

  region = data.gamefabric_region.europe.name
  replicas = [
    {
      region_type  = "baremetal"
      min_replicas = 10
      max_replicas = 200
      buffer_size  = 10
    },
    {
      region_type  = "cloud"
      min_replicas = 5
      max_replicas = 100
      buffer_size  = 5
    }
  ]
  containers = [
    {
      name  = "default" # the gameserver container should always be named "default"
      image = data.gamefabric_image.gameserver.image_target
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      envs = [
        {
          name  = "REGION"
          value = data.gamefabric_region.europe.name
        },
        {
          name = "REGION_TYPE"
          value_from = {
            field_path = "metadata.regionTypeName"
          }
        }
      ]
      ports = [
        {
          name           = "game"
          protocol       = "TCP"
          container_port = 27015
          policy         = "Dynamic"
        }
      ]
    }
  ]
}
```

## Example Advanced Usage - Armada for All Regions with Dynamic Sizing

This example sets up an armada in all available regions, it uses the result of the `gamefabric_regions` data source to create an armada in each region within the environment.
It also calculates the minimum, maximum and buffer sizes based on available CPU/Memory resources within a region type.

This allows the armada replicas to be adjusted by simply changing the game_cpu_request and game_memory_request locals.

```terraform
data "gamefabric_branch" "prod" {
  display_name = "prod"
}

data "gamefabric_image" "gameserver" {
  branch = data.gamefabric_branch.prod.name
  image  = "gameserver"
  tag    = "1.2.3"
}

data "gamefabric_environment" "prod" {
  name = "prod"
}

data "gamefabric_regions" "all" {
  environment = data.gamefabric_environment.prod.name
}

locals {
  game_cpu_request    = "250m"
  game_memory_request = "250Mi"

  # Normalize CPU request to cores
  game_cpu_normalized = endswith(local.game_cpu_request, "m") ? coalesce(tonumber(trimsuffix(local.game_cpu_request, "m")) / 1000, 0) : coalesce(tonumber(local.game_cpu_request), 0)

  # Normalize memory request to Gi
  game_memory_normalized = (
    endswith(local.game_memory_request, "Ki") ?
    coalesce(tonumber(trimsuffix(local.game_memory_request, "Ki")), 0) / 1024 / 1024 :
    (
      endswith(local.game_memory_request, "Mi") ?
      coalesce(tonumber(trimsuffix(local.game_memory_request, "Mi")), 0) / 1024 :
      (
        endswith(local.game_memory_request, "Gi") ?
        coalesce(tonumber(trimsuffix(local.game_memory_request, "Gi")), 0) :
        coalesce(0)
      )
    )
  )

  # Map of all regions from the data source
  regions = { for index, region in data.gamefabric_regions.all.regions : region.name => region }

  # Extract types and their resources for each region
  types_per_region = {
    for region_name, region in local.regions :
    region_name => { for type in keys(region.types) : type => region.types[type] }
  }

  # Normalize region resources to cores and Gi
  normalized_types_per_region = {
    for region_name, types in local.types_per_region :
    region_name => {
      for type_name, type_resources in types :
      type_name => {
        # Normalize CPU to cores
        cpu = endswith(type_resources.cpu, "m") ? coalesce(tonumber(trimsuffix(type_resources.cpu, "m")) / 1000, 0) : coalesce(tonumber(type_resources.cpu), 0)
        # Normalize memory to Gi
        memory = (
          endswith(type_resources.memory, "Ki") ?
          coalesce(tonumber(trimsuffix(type_resources.memory, "Ki")), 0) / 1024 / 1024 :
          (
            endswith(type_resources.memory, "Mi") ?
            coalesce(tonumber(trimsuffix(type_resources.memory, "Mi")), 0) / 1024 :
            (
              endswith(type_resources.memory, "Gi") ?
              coalesce(tonumber(trimsuffix(type_resources.memory, "Gi")), 0) :
              coalesce(0)
            )
          )
        )
      }
    }
  }

  # Calculate replicas for each region and type
  replicas_per_region = {
    for region_name, types in local.normalized_types_per_region :
    region_name => [
      for type_name, type_resources in types : {
        region_type = type_name
        max_replicas = floor(min(
          # CPU capacity divided by CPU request (both normalized to cores)
          type_resources.cpu / local.game_cpu_normalized,
          # Memory capacity divided by memory request (both normalized to Gi)
          type_resources.memory / local.game_memory_normalized
        ))
        min_replicas = floor(min(
          type_resources.cpu / local.game_cpu_normalized,
          type_resources.memory / local.game_memory_normalized
          )) == 0 ? 0 : max(1, floor(floor(min(
            type_resources.cpu / local.game_cpu_normalized,
            type_resources.memory / local.game_memory_normalized
        )) * 0.1)) # 10% of max_replicas, at least 1 if possible
        buffer_size = floor(min(
          type_resources.cpu / local.game_cpu_normalized,
          type_resources.memory / local.game_memory_normalized
          )) == 0 ? 0 : max(1, floor(floor(min(
            type_resources.cpu / local.game_cpu_normalized,
            type_resources.memory / local.game_memory_normalized
        )) * 0.05)) # 5% of max_replicas, at least 1 if possible
      }
    ]
  }
}

resource "gamefabric_armada" "this" {
  for_each    = local.regions
  name        = "myarmada-${each.key}"
  environment = data.gamefabric_environment.prod.name

  region   = each.key
  replicas = local.replicas_per_region[each.key]
  containers = [
    {
      name  = "default" # the gameserver container should always be named "default"
      image = data.gamefabric_image.gameserver.image_target
      resources = {
        requests = {
          cpu    = local.game_cpu_request
          memory = local.game_memory_request
        }
      }
      envs = [
        {
          name  = "REGION"
          value = data.gamefabric_region.europe.name
        },
        {
          name = "REGION_TYPE"
          value_from = {
            field_path = "metadata.regionTypeName"
          }
        }
      ]
      ports = [
        {
          name           = "game"
          protocol       = "TCP"
          container_port = 27015
          policy         = "Dynamic"
        }
      ]
    }
  ]
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `containers` (Attributes List) Containers is a list of containers belonging to the game server. (see [below for nested schema](#nestedatt--containers))
- `environment` (String) The name of the environment the resource belongs to.
- `name` (String) The unique object name within its scope. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 49 characters.
- `region` (String) Region defines the region the game servers are distributed to.

### Optional

- `annotations` (Map of String) Annotations is an unstructured map of keys and values stored on an object.
- `autoscaling` (Attributes) Autoscaling configuration for the game servers. (see [below for nested schema](#nestedatt--autoscaling))
- `description` (String) Description is the optional description of the armada.
- `gameserver_annotations` (Map of String) Annotations for the game server pods.
- `gameserver_labels` (Map of String) Labels for the game server pods.
- `gateway_policies` (List of String) GatewayPolicies is a list of gateway policies to apply to the Armada.
- `health_checks` (Attributes) HealthChecks is the health checking configuration for Agones game servers. (see [below for nested schema](#nestedatt--health_checks))
- `labels` (Map of String) A map of keys and values that can be used to organize and categorize objects.
- `profiling_enabled` (Boolean) ProfilingEnabled indicates whether profiling is enabled for the Armada.
- `replicas` (Attributes List) A replicas specifies the distribution of game servers across the available types of capacity in the selected region type. (see [below for nested schema](#nestedatt--replicas))
- `strategy` (Attributes) Strategy defines the rollout strategy for updating game servers. The default is RollingUpdate. (see [below for nested schema](#nestedatt--strategy))
- `termination_configuration` (Attributes) TerminationConfiguration defines the termination grace period for game servers. (see [below for nested schema](#nestedatt--termination_configuration))
- `volumes` (Attributes List) Volumes is a list of volumes that can be mounted by containers belonging to the game server. (see [below for nested schema](#nestedatt--volumes))

### Read-Only

- `id` (String) The unique Terraform identifier.
- `image_updater_target` (Attributes) ImageUpdaterTarget is the reference that an image updater can target to match the Armada. (see [below for nested schema](#nestedatt--image_updater_target))

<a id="nestedatt--containers"></a>
### Nested Schema for `containers`

Required:

- `image` (Attributes) Image is the GameFabric container image to run. You can use the `image_target` attribute of the `gamefabric_image` datasource to set this. (see [below for nested schema](#nestedatt--containers--image))
- `name` (String) Name is the name of the container. The primary gameserver container should be named `default`

Optional:

- `args` (List of String) Args are arguments to the entrypoint.
- `command` (List of String) Command is the entrypoint array. This is not executed within a shell.
- `config_files` (Attributes List) ConfigFiles is a list of configuration files to mount into the containers filesystem. (see [below for nested schema](#nestedatt--containers--config_files))
- `envs` (Attributes List) Envs is a list of environment variables to set on all containers in this Armada. (see [below for nested schema](#nestedatt--containers--envs))
- `ports` (Attributes List) Ports are the ports to expose from the container. (see [below for nested schema](#nestedatt--containers--ports))
- `resources` (Attributes) Resources describes the compute resource requirements. See the <a href="https://docs.gamefabric.com/multiplayer-servers/multiplayer-services/resource-management">GameFabric documentation</a> for more details on how to configure resource requests and limits. (see [below for nested schema](#nestedatt--containers--resources))
- `volume_mounts` (Attributes List) VolumeMounts are the volumes to mount into the container&#39;s filesystem. (see [below for nested schema](#nestedatt--containers--volume_mounts))

<a id="nestedatt--containers--image"></a>
### Nested Schema for `containers.image`

Required:

- `branch` (String) Branch of the GameFabric image.
- `name` (String) Name is the name of the GameFabric image.


<a id="nestedatt--containers--config_files"></a>
### Nested Schema for `containers.config_files`

Required:

- `mount_path` (String) MountPath is the path to mount the configuration file on.
- `name` (String) Name is the name of the configuration file.


<a id="nestedatt--containers--envs"></a>
### Nested Schema for `containers.envs`

Required:

- `name` (String) Name is the name of the environment variable.

Optional:

- `value` (String) Value is the value of the environment variable.
- `value_from` (Attributes) ValueFrom is the source for the environment variable&#39;s value. (see [below for nested schema](#nestedatt--containers--envs--value_from))

<a id="nestedatt--containers--envs--value_from"></a>
### Nested Schema for `containers.envs.value_from`

Optional:

- `config_file` (String) ConfigFile select the configuration file.
- `field_path` (String) FieldPath selects the field of the pod. Supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, metadata.armadaName, metadata.regionName, metadata.regionTypeName, metadata.siteName, metadata.imageBranch, metadata.imageName, metadata.imageTag, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.



<a id="nestedatt--containers--ports"></a>
### Nested Schema for `containers.ports`

Required:

- `name` (String) Name is the name of the port. Must contain only lowercase alphanumeric characters, hyphens, or dots. Must start and end with an alphanumeric character. Maximum length is 63 characters.
- `policy` (String) Policy defines how the host port is populated. `Dynamic` (default) allocates a free host port and maps it to the `container_port` (required). The gameserver must report the external port (obtained via Agones SDK) to backends for client connections. `Passthrough` dynamically allocates a host port and sets `container_port` to match it. The gameserver must discover this port via Agones SDK and listen on it.

Optional:

- `container_port` (Number) ContainerPort is the port that is being opened on the specified container&#39;s process.
- `protection_protocol` (String) ProtectionProtocol is the optional name of the protection protocol being used.
- `protocol` (String) Protocol is the network protocol being used. Defaults to UDP. TCP is the other option.


<a id="nestedatt--containers--resources"></a>
### Nested Schema for `containers.resources`

Optional:

- `limits` (Attributes) Limits describes the maximum amount of compute resources allowed. (see [below for nested schema](#nestedatt--containers--resources--limits))
- `requests` (Attributes) Requests describes the minimum amount of compute resources required. (see [below for nested schema](#nestedatt--containers--resources--requests))

<a id="nestedatt--containers--resources--limits"></a>
### Nested Schema for `containers.resources.limits`

Optional:

- `cpu` (String) CPU limit.
- `memory` (String) Memory limit.


<a id="nestedatt--containers--resources--requests"></a>
### Nested Schema for `containers.resources.requests`

Optional:

- `cpu` (String) CPU request.
- `memory` (String) Memory request.



<a id="nestedatt--containers--volume_mounts"></a>
### Nested Schema for `containers.volume_mounts`

Optional:

- `mount_path` (String) Path within the container at which the volume should be mounted.
- `name` (String) Name is the name of the volume.
- `sub_path` (String) Path within the volume from which the container's volume should be mounted.
- `sub_path_expr` (String) Expanded path within the volume from which the container's volume should be mounted.



<a id="nestedatt--autoscaling"></a>
### Nested Schema for `autoscaling`

Optional:

- `fixed_interval_seconds` (Number) Defines how often the auto-scaler re-evaluates the number of game servers.


<a id="nestedatt--health_checks"></a>
### Nested Schema for `health_checks`

Optional:

- `disabled` (Boolean) Disabled indicates whether Agones health checks are disabled.
- `failure_threshold` (Number) FailureThreshold is the number of consecutive failures before the game server is marked unhealthy and terminated.
- `initial_delay_seconds` (Number) InitialDelaySeconds is the number of seconds to wait before performing the first check.
- `period_seconds` (Number) PeriodSeconds is the number of seconds between checks.


<a id="nestedatt--replicas"></a>
### Nested Schema for `replicas`

Required:

- `buffer_size` (Number) BufferSize is the number of replicas to have ready all the time.
- `max_replicas` (Number) MaxReplicas is the maximum number of replicas in the region type.
- `min_replicas` (Number) MinReplicas is the minimum number of replicas in the region type.
- `region_type` (String) RegionType is the name of the region type.


<a id="nestedatt--strategy"></a>
### Nested Schema for `strategy`

Optional:

- `recreate` (Attributes) Recreate defines the recreate strategy which will recreate all unallocated gameservers at once on updates. This should only be used for development workloads or where downtime is acceptable. (see [below for nested schema](#nestedatt--strategy--recreate))
- `rolling_update` (Attributes) RollingUpdate defines the rolling update strategy, which gradually replaces game servers with new ones. (see [below for nested schema](#nestedatt--strategy--rolling_update))

<a id="nestedatt--strategy--recreate"></a>
### Nested Schema for `strategy.recreate`


<a id="nestedatt--strategy--rolling_update"></a>
### Nested Schema for `strategy.rolling_update`

Optional:

- `max_surge` (String) MaxSurge is the maximum number of game servers that can be created over the desired number of game servers during an update.
- `max_unavailable` (String) MaxUnavailable is the maximum number of game servers that can be unavailable during an update.



<a id="nestedatt--termination_configuration"></a>
### Nested Schema for `termination_configuration`

Optional:

- `grace_period_seconds` (Number) GracePeriodSeconds is the duration in seconds the game server needs to terminate gracefully.


<a id="nestedatt--volumes"></a>
### Nested Schema for `volumes`

Required:

- `name` (String) Name is the name of the volume.

Optional:

- `empty_dir` (Attributes) EmptyDir represents a volume that is initially empty. (see [below for nested schema](#nestedatt--volumes--empty_dir))

<a id="nestedatt--volumes--empty_dir"></a>
### Nested Schema for `volumes.empty_dir`

Optional:

- `size_limit` (String) SizeLimit is the total amount of local storage required for this EmptyDir volume.



<a id="nestedatt--image_updater_target"></a>
### Nested Schema for `image_updater_target`

Read-Only:

- `environment` (String) Environment is the environment of the image updater target.
- `name` (String) Name is the name of the image updater target.
- `type` (String) Type is the type of the image updater target.

## Import

Import is supported using the following syntax:

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute, for example:

```terraform
import {
  id = "{{ environment }}/{{ name }}"
  to = gamefabric_armada.europe
}
```

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import gamefabric_armada.europe "{{ environment }}/{{ name }}"
```