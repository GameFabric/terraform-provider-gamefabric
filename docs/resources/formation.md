---
page_title: "gamefabric_formation Resource - GameFabric"
subcategory: ""
description: |-
  
---

# gamefabric_formation (Resource)

A Formation is used to create and manage a set of vessels with shared configuration that is applied to all vessels of the formation, while also supporting per-vessel overrides for specific attributes.

Formations simplify the management of multiple related vessels by centralizing their common configuration, making it easier to maintain consistency across your dedicated servers while still allowing customization where needed.

At scale, formations provide significant benefits over managing individual vessels:
- **Terraform State Efficiency**: A formation with 400 vessels counts as a single resource in the Terraform state, compared to 400 separate resources when using individual vessel resources. This dramatically reduces state file size and improves Terraform operations performance.
- **Unified Overview**: The GameFabric frontend provides a consolidated view of all vessels within a formation, making it easier to monitor and manage large deployments.

For details check the <a href="https://docs.gamefabric.com/multiplayer-servers/getting-started/glossary#formation">GameFabric documentation</a>.


## Example Usage - Simple Formation

This example sets up a minimal formation with vessels.

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

resource "gamefabric_formation" "this" {
  name        = "my-formation"
  environment = data.gamefabric_environment.prod.name

  vessels = [
    {
      name = "vessel-1"
      region = data.gamefabric_region.europe.name
    }
  ]

  containers = [
    {
      name      = "default" # the gameserver container should always be named "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      envs = [
        {
          name  = "SERVER_NAME"
          value = "My Server ${data.gamefabric_region.europe.display_name}"
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

## Advanced Usage - Formation with Multiple Game Modes

This example demonstrates how to create a formation with multiple vessels for different game modes.
It will create one formation per game mode with 3 vessels within the europe region and configure the game modes and server names via environment variables.
The formations are named `survival`, `creative`, and `hardcore` and the server names are "Survival 1 - EU", "Creative 1 - EU" etc.

```terraform
locals {
  game_modes = {
    survival = {
      display_name = "Survival"
      mode_value   = "survival"
    }
    creative = {
      display_name = "Creative"
      mode_value   = "creative"
    }
    hardcore = {
      display_name = "Hardcore"
      mode_value   = "hardcore"
    }
  }

  # Create vessels configuration for each game mode
  vessels_per_mode = {
    for mode_key, mode in local.game_modes : mode_key => [
      for i in range(3) : {
        name        = "${mode_key}-${i}"
        region      = data.gamefabric_region.europe.name
        description = "${mode.display_name} server ${i + 1}"
        suspend     = false
        override = {
          containers = [
            {
              envs = [
                {
                  name  = "SERVER_NAME"
                  value = "${mode.display_name} ${i + 1} - EU"
                }
              ]
            }
          ]
        }
      }
    ]
  }
}

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

resource "gamefabric_formation" "survival" {
  for_each    = local.game_modes
  name        = each.key
  environment = data.gamefabric_environment.prod.name

  vessels = local.vessels_per_mode[each.key]

  // Containers
  containers = [
    {
      name      = "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
      envs = [
        {
          name  = "GAME_MODE"
          value = each.value.mode_value
        }
      ]
      ports = [
        {
          name           = "http"
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
- `environment` (String) The name of the environment the object belongs to.
- `name` (String) The unique object name within its scope.
- `vessels` (Attributes List) Vessels is a list of vessels belonging to the game server. (see [below for nested schema](#nestedatt--vessels))

### Optional

- `annotations` (Map of String) Annotations is an unstructured map of keys and values stored on an object.
- `description` (String) Description is the optional description of the Formation.
- `gameserver_annotations` (Map of String) Annotations is an unstructured map of keys and values stored on an object.
- `gameserver_labels` (Map of String) A map of keys and values that can be used to organize and categorize objects.
- `gateway_policies` (List of String) GatewayPolicies is a list of gateway policies to apply to the Formation.
- `health_checks` (Attributes) HealthChecks is the health checking configuration for Agones game servers. (see [below for nested schema](#nestedatt--health_checks))
- `labels` (Map of String) A map of keys and values that can be used to organize and categorize objects.
- `profiling_enabled` (Boolean) ProfilingEnabled indicates whether profiling is enabled for the Formation.
- `termination_configuration` (Attributes) TerminationConfiguration defines the termination grace period for game servers. (see [below for nested schema](#nestedatt--termination_configuration))
- `volume_templates` (Attributes List) VolumeTemplates are the templates for volumes that can be mounted by containers belonging to the game server. (see [below for nested schema](#nestedatt--volume_templates))
- `volumes` (Attributes List) Volumes is a list of volumes that can be mounted by containers belonging to the game server. (see [below for nested schema](#nestedatt--volumes))

### Read-Only

- `id` (String) The unique Terraform identifier.
- `image_updater_target` (Attributes) ImageUpdaterTarget is the reference that an image updater can target to match the Formation. (see [below for nested schema](#nestedatt--image_updater_target))

<a id="nestedatt--containers"></a>
### Nested Schema for `containers`

Required:

- `image_ref` (Attributes) The reference to the image to run your container with. You should use the `image_ref` attribute of the `gamefabric_image` datasource to configure this attribute. (see [below for nested schema](#nestedatt--containers--image_ref))
- `name` (String) Name is the name of the container. The primary gameserver container should be named `default`

Optional:

- `args` (List of String) Args are arguments to the entrypoint.
- `command` (List of String) Command is the entrypoint array. This is not executed within a shell.
- `config_files` (Attributes List) ConfigFiles is a list of configuration files to mount into the containers filesystem. (see [below for nested schema](#nestedatt--containers--config_files))
- `envs` (Attributes List) Envs is a list of environment variables to set on all containers in this Armada. (see [below for nested schema](#nestedatt--containers--envs))
- `ports` (Attributes List) Ports are the ports to expose from the container. (see [below for nested schema](#nestedatt--containers--ports))
- `resources` (Attributes) Resources describes the compute resource requirements. See the <a href="https://docs.gamefabric.com/multiplayer-servers/multiplayer-services/resource-management">GameFabric documentation</a> for more details on how to configure resource requests and limits. (see [below for nested schema](#nestedatt--containers--resources))
- `volume_mounts` (Attributes List) VolumeMounts are the volumes to mount into the container&#39;s filesystem. (see [below for nested schema](#nestedatt--containers--volume_mounts))

<a id="nestedatt--containers--image_ref"></a>
### Nested Schema for `containers.image_ref`

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
- `field_path` (String) FieldPath selects the field of the pod. Supports `metadata.name`, `metadata.namespace`, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, `metadata.armadaName`, `metadata.regionName`, `metadata.regionTypeName`, `metadata.siteName`, `metadata.imageBranch`, `metadata.imageName`, `metadata.imageTag`, `spec.nodeName`, `spec.serviceAccountName`, `status.hostIP`, `status.podIP`, `status.podIPs`.



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



<a id="nestedatt--vessels"></a>
### Nested Schema for `vessels`

Required:

- `name` (String) Name is the name of the vessel.
- `region` (String) Region defines the region the game servers are distributed to.

Optional:

- `description` (String) Description is the optional description of the Vessel.
- `override` (Attributes) Override is the optional override configuration for Vessels. (see [below for nested schema](#nestedatt--vessels--override))
- `suspend` (Boolean) Suspend indicates whether the Vessel should be suspended.

<a id="nestedatt--vessels--override"></a>
### Nested Schema for `vessels.override`

Optional:

- `containers` (Attributes List) Containers is a list of containers belonging to the game server. (see [below for nested schema](#nestedatt--vessels--override--containers))
- `gameserver_labels` (Map of String) A map of keys and values that can be used to organize and categorize objects.

<a id="nestedatt--vessels--override--containers"></a>
### Nested Schema for `vessels.override.containers`

Optional:

- `args` (List of String) Args are arguments to the entrypoint.
- `command` (List of String) Command is the entrypoint array. This is not executed within a shell.
- `envs` (Attributes List) Envs is a list of environment variables to set on all containers in this Armada. (see [below for nested schema](#nestedatt--vessels--override--containers--envs))

<a id="nestedatt--vessels--override--containers--envs"></a>
### Nested Schema for `vessels.override.containers.envs`

Required:

- `name` (String) Name is the name of the environment variable.

Optional:

- `value` (String) Value is the value of the environment variable.
- `value_from` (Attributes) ValueFrom is the source for the environment variable&#39;s value. (see [below for nested schema](#nestedatt--vessels--override--containers--envs--value_from))

<a id="nestedatt--vessels--override--containers--envs--value_from"></a>
### Nested Schema for `vessels.override.containers.envs.value_from`

Optional:

- `config_file` (String) ConfigFile select the configuration file.
- `field_path` (String) FieldPath selects the field of the pod. Supports `metadata.name`, `metadata.namespace`, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, `metadata.armadaName`, `metadata.regionName`, `metadata.regionTypeName`, `metadata.siteName`, `metadata.imageBranch`, `metadata.imageName`, `metadata.imageTag`, `spec.nodeName`, `spec.serviceAccountName`, `status.hostIP`, `status.podIP`, `status.podIPs`.






<a id="nestedatt--health_checks"></a>
### Nested Schema for `health_checks`

Optional:

- `disabled` (Boolean) Disabled indicates whether Agones health checks are disabled.
- `failure_threshold` (Number) FailureThreshold is the number of consecutive failures before the game server is marked unhealthy and terminated.
- `initial_delay_seconds` (Number) InitialDelaySeconds is the number of seconds to wait before performing the first check.
- `period_seconds` (Number) PeriodSeconds is the number of seconds between checks.


<a id="nestedatt--termination_configuration"></a>
### Nested Schema for `termination_configuration`

Optional:

- `grace_period_seconds` (Number) The duration in seconds the game server needs to terminate gracefully.
- `maintenance_seconds` (Number) The duration in seconds the game server needs to terminate gracefully when put into maintenance.
- `spec_change_seconds` (Number) The duration in seconds the game server needs to terminate gracefully after Formation specs have changed.
- `user_initiated_seconds` (Number) The duration in seconds the game server needs to terminate gracefully after a user has initiated a termination.


<a id="nestedatt--volume_templates"></a>
### Nested Schema for `volume_templates`

Required:

- `capacity` (String) Capacity is the storage capacity requested for the volume.
- `name` (String) Name is the name of the volume as it will be referenced.
- `reclaim_policy` (String) The reclaim policy for a volume created by a Formation. Valid values are `Delete` (recommended) and `Retain`. `Delete` means that a backing Kubernetes volume is automatically deleted when the user deletes the persistent volume. `Retain` means if a user deletes the persistent volume, the backing Kubernetes volume will not be deleted. Instead, it is moved to the Released phase, where all of its data can be manually recovered by Nitrado for an undetermined time.
- `volume_store_name` (String) Name is the name of the VolumeStore.


<a id="nestedatt--volumes"></a>
### Nested Schema for `volumes`

Required:

- `name` (String) Name is the name of the volume.

Optional:

- `empty_dir` (Attributes) EmptyDir represents a volume that is initially empty. (see [below for nested schema](#nestedatt--volumes--empty_dir))
- `persistent` (Attributes) Persistent represents a volume with data persistence. (see [below for nested schema](#nestedatt--volumes--persistent))

<a id="nestedatt--volumes--empty_dir"></a>
### Nested Schema for `volumes.empty_dir`

Optional:

- `size_limit` (String) SizeLimit is the total amount of local storage required for this EmptyDir volume.


<a id="nestedatt--volumes--persistent"></a>
### Nested Schema for `volumes.persistent`

Optional:

- `volume_name` (String) VolumeName is the name of the persistent volume.



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
  to = gamefabric_formation.my_server
}
```

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import gamefabric_formation.my_server "{{ environment }}/{{ name }}"
```
