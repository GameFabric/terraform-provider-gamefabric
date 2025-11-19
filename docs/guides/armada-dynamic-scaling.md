---
page_title: "Dynamic Armada Scaling Across All Regions"
subcategory: "Guides"
description: |-
  Learn how to create armadas across all regions with dynamic replica sizing based on available resources.
---

# Dynamic Armada Scaling Across All Regions

This guide demonstrates how to create armadas across all available regions in an environment with dynamic replica sizing. The configuration automatically calculates minimum replicas, maximum replicas, and buffer sizes based on the available CPU and memory resources within each region type.

## Overview

This approach is useful when you want to:

- Deploy game servers to all regions in your environment
- Automatically scale based on available infrastructure resources
- Maintain consistent resource allocation across different region types
- Adjust capacity by simply changing CPU/memory request values

## Prerequisites

Before following this guide, you should have:

- A GameFabric environment configured
- A container branch with your game server image
- Knowledge of your game server's resource requirements

## Complete Example

The following example creates an armada in each available region, automatically calculating replica counts based on the resources available in each region type:

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
      name  = "default" # the game server container should always be named "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      resources = {
        requests = {
          cpu    = local.game_cpu_request
          memory = local.game_memory_request
        }
      }
      envs = [
        {
          name  = "REGION"
          value = each.key
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

## How It Works

### 1. Resource Requirements

First, define your game server's resource requirements:

```terraform
locals {
  game_cpu_request    = "250m"
  game_memory_request = "250Mi"
}
```

These values determine how many game servers can fit on each type of infrastructure.

### 2. Resource Normalization

The configuration normalizes all resource values to standard units (CPU cores and GiB of memory). This ensures consistent calculations regardless of the unit format (e.g., "250m", "0.25", "250Mi", "0.25Gi").

### 3. Region Data Extraction

The `gamefabric_regions` data source provides information about all available regions and their resource capacity by type (e.g., baremetal, cloud). The configuration processes this data to create a normalized view of available resources.

### 4. Replica Calculation

For each region and type, the configuration calculates:

- **max_replicas**: Maximum number of game servers that can run based on available resources
  - Calculated as: `min(available_cpu / requested_cpu, available_memory / requested_memory)`
- **min_replicas**: Baseline capacity (10% of max_replicas, minimum of 1)
  - Ensures there's always some capacity available for quick scaling
- **buffer_size**: Extra servers kept ready (5% of max_replicas, minimum of 1)
  - Maintains a pool of ready servers for instant player connections

### 5. Dynamic Armada Creation

The `for_each` on the `gamefabric_armada` resource creates one armada per region, each configured with the calculated replica settings for that region's available resources.

## Customization

### Adjusting Scaling Percentages

You can modify the min_replicas and buffer_size percentages to fit your needs:

```terraform
min_replicas = max(1, floor(max_replicas * 0.2))  # 20% instead of 10%
buffer_size = max(2, floor(max_replicas * 0.1))   # 10% instead of 5%, minimum of 2
```

### Filtering Regions

To deploy to specific regions only:

```terraform
locals {
  regions = { 
    for index, region in data.gamefabric_regions.all.regions : 
    region.name => region 
    if contains(["europe", "us-east"], region.name)
  }
}
```

### Different Resources Per Region

You can define region-specific resource requirements:

```terraform
locals {
  game_cpu_request = lookup({
    europe  = "500m"
    us-east = "250m"
  }, each.key, "250m")
}
```

## Best Practices

1. **Start Conservative**: Begin with lower resource requests and gradually increase based on actual usage
2. **Monitor Utilization**: Use GameFabric's monitoring to ensure calculated replica counts match your needs
3. **Test Scaling**: Validate that the min/max/buffer calculations work for your traffic patterns
4. **Version Control**: Keep this configuration in version control to track changes to scaling parameters
5. **Environment Separation**: Use different configurations for dev/staging/production environments

## Related Resources

- [gamefabric_armada](../resources/armada.md) - Armada resource documentation
- [gamefabric_regions](../data-sources/regions.md) - Regions data source documentation
- [gamefabric_image](../data-sources/image.md) - Image data source documentation
