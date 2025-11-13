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
      name      = "default" # the gameserver container should always be named "default"
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
