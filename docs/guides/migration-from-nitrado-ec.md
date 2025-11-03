---
page_title: "Migrating from Nitrado EC Provider"
subcategory: "Guides"
description: |-
  Learn how to migrate your infrastructure from the legacy Nitrado EC Terraform provider to the GameFabric provider.
---

# Migrating from Nitrado EC Provider to GameFabric Provider

This guide helps you migrate your existing infrastructure managed by the legacy [Nitrado EC Terraform provider](https://registry.terraform.io/providers/nitrado/ec/latest/docs) to the new GameFabric provider.

> **Disclaimer:** This migration guide was written with AI support and has been reviewed for accuracy. However, errors may still occur. If you encounter issues or discrepancies, please refer to the [official documentation for individual resources](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs) in the GameFabric provider. If the documentation doesn't resolve your issue, please reach out to our Customer Success Management team for assistance.

## Important Migration Considerations

⚠️ **In-place provider migration is not supported.** The GameFabric provider uses different resource names, attribute names, and structures compared to the legacy Nitrado EC provider. A direct provider swap will result in errors and potential infrastructure recreation.

### Migration Strategy

The recommended migration approach is to create a completely new Terraform workspace with a separate state file:

1. **Create a new workspace** with the GameFabric provider
2. **Write matching Terraform code** using the new provider syntax
3. **Import existing resources** into the new Terraform state using import blocks
4. **Verify the configuration** matches your existing infrastructure
5. **Keep the old workspace** as a backup and safety net

**Advantages of this approach:**
- Less error-prone and lower risk
- Old configuration remains functional if issues arise with the new provider
- Clean separation allows easier rollback
- No risk of accidentally destroying resources by removing old provider code
- Recommended for production environments

This approach ensures zero downtime and maintains your existing infrastructure throughout the migration.

## Prerequisites

Before starting the migration:

- Terraform v1.5.0 or later (required for import blocks)
- Access to your GameFabric installation
- A service account with appropriate permissions
- Understanding of your current infrastructure setup

## Step 1: Create New Workspace and Install GameFabric Provider

Create a new directory for your new Terraform workspace:

```bash
# Create new workspace directory
mkdir terraform-gamefabric
cd terraform-gamefabric
```

Create your `main.tf` or `provider.tf` with the GameFabric provider:

```terraform
terraform {
  required_providers {
    gamefabric = {
      source  = "gamefabric/gamefabric"
      version = "~> 0.3.0"
    }
  }
}

provider "gamefabric" {
  customer_id     = var.customer_id
  service_account = var.service_account
  password        = var.password  # or use GAMEFABRIC_PASSWORD env var
}
```

Run `terraform init` to download the provider.

## Step 2: Resource Mapping

Below is a mapping of resources from the Nitrado EC provider to the GameFabric provider:

| Nitrado EC Resource | GameFabric Resource | Notes |
|---------------------|---------------------|-------|
| `ec_core_environment` | `gamefabric_environment` | Attribute names changed, flattened structure, `metadata.labels` → `labels` |
| `ec_container_branch` | `gamefabric_branch` | Retention policies restructured, flattened structure |
| `ec_core_region` | `gamefabric_region` | Region types configuration changed, flattened structure |
| `ec_armada_armada` | `gamefabric_armada` | Significant restructuring of containers, replicas, ports, and labels |
| `ec_armada_armadaset` | `gamefabric_armadaset` | Similar to armada changes |
| `ec_protection_gatewaypolicy` | `gamefabric_protection_gatewaypolicy` | Flattened structure |
| N/A | `gamefabric_configfile` | **New resource** - ConfigFiles are now manageable resources |
| N/A | `gamefabric_imageupdater` | **New resource** - ImageUpdaters are now manageable resources |

**Note:** The old EC provider also had additional resources like `ec_formation_vessel`, `ec_formation_formation`, `ec_authentication_provider`, `ec_authentication_serviceaccount`, `ec_storage_volume`, `ec_storage_volumestore`, and `ec_protection_protocol` which are not covered in this guide. Support for formations and vessels will be added in future releases of the GameFabric provider. If you're using other resources, please consult the GameFabric documentation for equivalent resources.

### Data Sources Mapping

| Nitrado EC Data Source | GameFabric Data Source |
|------------------------|------------------------|
| `ec_core_environment` | `gamefabric_environment` / `gamefabric_environments` |
| `ec_container_branch` | `gamefabric_branch` / `gamefabric_branches` |
| `ec_core_region` | `gamefabric_region` / `gamefabric_regions` |
| `ec_core_location` | `gamefabric_location` / `gamefabric_locations` |
| `ec_container_image` | `gamefabric_image` |
| `ec_armada_armada` | Data source not available in GameFabric (use resource instead) |
| `ec_armada_armadaset` | Data source not available in GameFabric (use resource instead) |
| `ec_protection_gatewaypolicy` | `gamefabric_protection_gatewaypolicy` / `gamefabric_protection_gatewaypolicies` |
| `ec_protection_protocol` | `gamefabric_protection_protocol` / `gamefabric_protection_protocols` |
| N/A | `gamefabric_configfile` / `gamefabric_configfiles` - **New data sources** |

## Step 3: Migrate Each Resource Type

For each resource type, you'll create the new resource definition and add an import block. The import block tells Terraform to import the existing resource instead of creating a new one.

### Migrating Environments

**Old (Nitrado EC):**
```terraform
resource "ec_core_environment" "prod" {
  metadata {
    name = "prod"
  }
  
  spec {
    display_name = "Production"
    description  = "Production environment"
  }
}
```

**New (GameFabric) with Import:**
```terraform
# Import block for the existing environment
import {
  id = "prod"
  to = gamefabric_environment.prod
}

# New resource definition
resource "gamefabric_environment" "prod" {
  name         = "prod"
  display_name = "Production"
  description  = "Production environment"
  
}
```

**Key Changes:**
- Removed `metadata` and `spec` wrappers - attributes are now at the top level
- `metadata.name` → `name`
- `spec.display_name` → `display_name`
- `spec.description` → `description`
- `spec.labels` → `labels`
- Import ID format: `"<name>"`

### Migrating Branches

**Old (Nitrado EC):**
```terraform
resource "ec_container_branch" "prod" {
  metadata {
    name = "prod"
  }
  
  spec {
    display_name = "Production Branch"
    description  = "Production image branch"
    
    retention_policies {
      metadata {
        name = "keep-releases"
      }
      spec {
        image_name_regex = ".*"
        tag_name_regex   = "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
        keep_count       = 10
        keep_days        = 30
      }
    }
  }
}
```

**New (GameFabric) with Import:**
```terraform
# Import block for the existing branch
import {
  id = "prod"
  to = gamefabric_branch.prod
}

# New resource definition
resource "gamefabric_branch" "prod" {
  name         = "prod"
  display_name = "Production Branch"
  description  = "Production image branch"
  
  retention_policies = [
    {
      name        = "keep-releases"
      image_regex = ".*"
      tag_regex   = "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
      keep_count  = 10
      keep_days   = 30
    }
  ]
}
```

**Key Changes:**
- Flattened structure: removed `metadata` and `spec` wrappers
- `metadata.name` → `name`
- `spec.display_name` → `display_name`
- `spec.retention_policies[].metadata.name` → `retention_policies[].name`
- `spec.retention_policies[].spec.image_name_regex` → `retention_policies[].image_regex`
- `spec.retention_policies[].spec.tag_name_regex` → `retention_policies[].tag_regex`
- Import ID format: `"<name>"`

### Migrating ConfigFiles

**ConfigFiles are a now supported in the GameFabric provider.** In the old Nitrado EC provider, configuration files were referenced in environment variables using `config_file_key_ref`, but they were not manageable Terraform resources.

ConfigFiles are now a resources that you can manage with Terraform:

**New (GameFabric):**
```terraform
# If you want to manage an existing configfile with Terraform, import it first
import {
  id = "prod/game-config"
  to = gamefabric_configfile.game_config
}

# Then create or manage the configfile resource
resource "gamefabric_configfile" "game_config" {
  name        = "game-config"
  environment = "prod"
  
  labels = {
    type = "config"
  }
  
  data = <<EOT
[GameSettings]
MaxPlayers=64
EOT
}

# Then reference it in your containers
resource "gamefabric_armada" "game" {
  # ... other configuration ...
  
  containers = [
    {
      name = "default"
      # ... other container config ...
      
      envs = [
        {
          name = "CONFIG_CONTENT"
          value_from = {
            config_file = gamefabric_configfile.game_config.name
          }
        }
      ]
    }
  ]
}
```

**Migration Notes:**
- Environment variables can still reference any existing config file, regardless of whether they're managed by Terraform
- If you have existing config files that you want to manage with Terraform, you can create `gamefabric_configfile` resources and import them into your state
- This is optional - you can continue referencing existing config files by name without managing them in Terraform

### Migrating Regions

**Old (Nitrado EC):**
```terraform
resource "ec_core_region" "europe" {
  metadata {
    name        = "europe"
    environment = "prod"
  }
  
  spec {
    display_name = "Europe"
    description  = "European region"
    
    types {
      name      = "baremetal"
      locations = ["frankfurt-baremetal", "amsterdam-baremetal"]
      
      template {
        scheduling = "Distributed"
      }
    }
    
    types {
      name      = "gcp"
      locations = ["frankfurt-gcp"]
      
      template {
        scheduling = "Packed"
      }
    }
  }
}
```

**New (GameFabric) with Import:**
```terraform
# Import block for the existing region
import {
  id = "prod/europe"
  to = gamefabric_region.europe
}

# New resource definition
resource "gamefabric_region" "europe" {
  name         = "europe"
  display_name = "Europe"
  environment  = "prod"
  description  = "European region"
  
  types = [
    {
      name       = "baremetal"
      locations = ["frankfurt-baremetal", "amsterdam-baremetal"]
      scheduling = "Distributed"
    },
    {
      name       = "gcp"
      locations = ["frankfurt-gcp"]
      scheduling = "Packed"
    }
  ]
}
```

**Key Changes:**
- Flattened structure: removed nested `metadata` and `spec` wrappers
- `metadata.name` → `name`
- `metadata.environment` → `environment`
- `spec.display_name` → `display_name`
- `spec.types[].template.scheduling` → `types[].scheduling`
- `types` is now a list instead of repeated blocks
- Import ID format: `"<environment>/<name>"`

**Note:** Both providers use `locations` - the old provider documentation examples incorrectly showed `sites`, but the actual implementation used `locations`.

### Migrating Armadas

This is one of the most complex migrations due to significant restructuring of the armada resource.

**Old (Nitrado EC):**
```terraform
resource "ec_armada_armada" "game" {
  metadata {
    name        = "gameserver"
    environment = "prod"
    
    labels = {
      app = "gameserver"
    }
    
    annotations = {
      description = "Game server deployment"
    }
  }
  
  spec {
    description = "Game server armada"
    region      = "europe"
    
    autoscaling {
      fixed_interval {
        seconds = 30
      }
    }
    
    distribution {
      name         = "baremetal"
      min_replicas = 10
      max_replicas = 100
      buffer_size  = 10
    }
    
    distribution {
      name         = "gcp"
      min_replicas = 5
      max_replicas = 50
      buffer_size  = 5
    }
    
    template {
      metadata {
        labels = {
          version = "v1.2.3"
          tier    = "production"
        }
        
        annotations = {
          build = "12345"
        }
      }
      
      spec {
        containers {
          name   = "default"
          branch = "prod"
          image  = "gameserver"
          
          resources {
            requests = {
              cpu    = "250m"
              memory = "256Mi"
            }
            limits = {
              cpu    = "1000m"
              memory = "512Mi"
            }
          }
          
          env {
            name  = "GAME_MODE"
            value = "production"
          }
          
          env {
            name = "CONFIG_CONTENT"
            value_from {
              config_file_key_ref {
                name = "game-config"
              }
            }
          }
          
          env {
            name = "POD_NAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }
          
          ports {
            name           = "game"
            protocol       = "UDP"
            container_port = 7777
            policy         = "Dynamic"
            
            protection_protocol {
              value = "game-udp"
            }
          }
          
          config_files {
            name       = "server-cfg"
            mount_path = "/config/server.cfg"
          }
          
          volume_mounts {
            name       = "data"
            mount_path = "/data"
          }
        }
        
        health {
          disabled              = false
          initial_delay_seconds = 10
          period_seconds        = 5
          failure_threshold     = 3
        }
        
        termination_grace_period_seconds {
          value = 60
        }
        
        strategy {
          type = "RollingUpdate"
          
          rolling_update {
            max_surge {
              value = 1
            }
            max_unavailable {
              value = 0
            }
          }
        }
        
        volumes {
          name   = "data"
          medium = "Memory"
          size_limit {
            value = "1Gi"
          }
        }
        
        gateway_policies = ["ddos-protection"]
      }
    }
  }
}
```

**New (GameFabric) with Import:**
```terraform
# Import block for the existing armada
import {
  id = "prod/gameserver"
  to = gamefabric_armada.game
}

# Fetch the image ref
data "gamefabric_image" "gameserver" {
  branch = "prod"
  image  = "gameserver"
  tag    = "latest"  # or specific tag
}

# New resource definition
resource "gamefabric_armada" "game" {
  name        = "gameserver"
  environment = "prod"
  region      = "europe"
  description = "Game server armada"
  
  # Labels at armada level (previously metadata.labels)
  labels = {
    app = "gameserver"
  }
  
  # Annotations at armada level (previously metadata.annotations)
  annotations = {
    description = "Game server deployment"
  }
  
  # GameServer labels (previously spec.template.metadata.labels)
  gameserver_labels = {
    version = "v1.2.3"
    tier    = "production"
  }
  
  # GameServer annotations (previously spec.template.metadata.annotations)
  gameserver_annotations = {
    build = "12345"
  }
  
  # Autoscaling configuration
  autoscaling = {
    fixed_interval_seconds = 30
  }
  
  # Distribution → Replicas
  replicas = [
    {
      region_type  = "baremetal"
      min_replicas = 10
      max_replicas = 100
      buffer_size  = 10
    },
    {
      region_type  = "gcp"
      min_replicas = 5
      max_replicas = 50
      buffer_size  = 5
    }
  ]
  
  containers = [
    {
      name      = "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }
      
      envs = [
        {
          name  = "GAME_MODE"
          value = "production"
        },
        {
          name = "CONFIG_CONTENT"
          value_from = {
            config_file = "game-config"
          }
        },
        {
          name = "POD_NAME"
          value_from = {
            field_path = "metadata.name"
          }
        }
      ]
      
      ports = [
        {
          name                = "game"
          protocol            = "UDP"
          container_port      = 7777
          policy              = "Dynamic"
          protection_protocol = "game-udp"
        }
      ]
      
      config_files = [
        {
          name       = "server-cfg"
          mount_path = "/config/server.cfg"
        }
      ]
      
      volume_mounts = [
        {
          name       = "data"
          mount_path = "/data"
        }
      ]
    }
  ]
  
  # Health check configuration
  health_checks = {
    disabled              = false
    initial_delay_seconds = 10
    period_seconds        = 5
    failure_threshold     = 3
  }
  
  # Termination configuration
  termination_configuration = {
    grace_period_seconds = 60
  }
  
  # Rollout strategy
  strategy = {
    rolling_update = {
      max_surge       = 1
      max_unavailable = 0
    }
  }
  
  # Volumes
  volumes = [
    {
      name = "data"
      empty_dir = {
        size_limit = "1Gi"
      }
    }
  ]
  
  # Gateway policies
  gateway_policies = ["ddos-protection"]
}
```

**Key Changes:**
- **Flattened structure**: Removed `metadata` and `spec` wrappers throughout
- **Top-level attributes**:
  - `metadata.name` → `name`
  - `metadata.environment` → `environment`
  - `metadata.labels` → `labels` (armada-level labels)
  - `metadata.annotations` → `annotations` (armada-level annotations)
  - `spec.template.metadata.labels` → `gameserver_labels` (applied to game server pods)
  - `spec.template.metadata.annotations` → `gameserver_annotations` (applied to game server pods)
  - `spec.region` → `region`
  - `spec.description` → `description`
- **Autoscaling**:
  - `spec.autoscaling.fixed_interval.seconds` → `autoscaling.fixed_interval_seconds` (now an object)
- **Distribution → Replicas**:
  - Changed from repeated `distribution` blocks to `replicas` list
  - `distribution.name` → `replicas[].region_type`
  - Field names stay the same: `min_replicas`, `max_replicas`, `buffer_size`
- **Containers** (`spec.template.spec.containers` → `containers`):
  - `containers[].branch` and `containers[].image` → `containers[].image_ref` (use data source)
  - `containers[].env` → `containers[].envs`
  - `containers[].env[].value_from.config_file_key_ref.name` → `containers[].envs[].value_from.config_file`
  - `containers[].env[].value_from.field_ref.field_path` → `containers[].envs[].value_from.field_path` (simplified)
  - `containers[].ports[].protection_protocol.value` → `containers[].ports[].protection_protocol` (direct string)
  - All nested lists (envs, ports, config_files, volume_mounts) use list syntax
- **Health checks**:
  - `spec.template.spec.health` → `health_checks` (renamed to `health_checks`)
- **Termination grace period**:
  - `spec.template.spec.termination_grace_period_seconds.value` → `termination_configuration.grace_period_seconds` (now an object)
- **Strategy**:
  - `spec.template.spec.strategy` → `strategy`
  - `strategy.type` field removed - type is inferred from which nested block is used (`rolling_update` or `recreate`)
  - `strategy.rolling_update.max_surge.value` → `strategy.rolling_update.max_surge` (direct value - string or integer)
  - `strategy.rolling_update.max_unavailable.value` → `strategy.rolling_update.max_unavailable` (direct value - string or integer)
- **Volumes**:
  - `spec.template.spec.volumes` → `volumes` (list)
  - `volumes[].medium` and `volumes[].size_limit.value` → `volumes[].empty_dir.size_limit` (restructured as typed volume)
- **Gateway policies**:
  - `spec.template.spec.gateway_policies` → `gateway_policies` (stays the same)
- **Import ID format**: `"<environment>/<name>"`

### Migrating ArmadaSets

ArmadaSets are even more complex than Armadas as they manage multiple armadas across different regions. The structure has significant changes between the old and new provider.

**Old (Nitrado EC):**
```terraform
resource "ec_armada_armadaset" "game_set" {
  metadata {
    name        = "gameserver-set"
    environment = "prod"
    
    labels = {
      app     = "gameserver"
      managed = "terraform"
    }
    
    annotations = {
      owner = "platform-team"
    }
  }
  
  spec {
    description = "Multi-region game server deployment"
    
    autoscaling {
      fixed_interval {
        seconds = 30
      }
    }
    
    # Define armadas for different regions
    armadas {
      region      = "europe"
      description = "European game servers"
      
      distribution {
        name         = "baremetal"
        min_replicas = 10
        max_replicas = 100
        buffer_size  = 10
      }
      
      distribution {
        name         = "gcp"
        min_replicas = 5
        max_replicas = 50
        buffer_size  = 5
      }
    }
    
    armadas {
      region      = "us-east"
      description = "US East game servers"
      
      distribution {
        name         = "aws"
        min_replicas = 8
        max_replicas = 80
        buffer_size  = 8
      }
    }
    
    # Optional: Override template settings for specific regions
    override {
      region = "us-east"
      
      labels = {
        datacenter = "us-east-1"
      }
      
      env {
        name  = "REGION"
        value = "us-east"
      }
      
      env {
        name  = "GAME_MODE"
        value = "ranked"  # Override the default "production" value
      }
    }
    
    # Template applies to all armadas
    template {
      metadata {
        labels = {
          version = "v1.2.3"
          tier    = "production"
        }
        
        annotations = {
          build = "12345"
        }
      }
      
      spec {
        containers {
          name   = "default"
          branch = "prod"
          image  = "gameserver"
          
          resources {
            requests = {
              cpu    = "250m"
              memory = "256Mi"
            }
            limits = {
              cpu    = "1000m"
              memory = "512Mi"
            }
          }
          
          env {
            name  = "GAME_MODE"
            value = "production"
          }
          
          env {
            name = "CONFIG_CONTENT"
            value_from {
              config_file_key_ref {
                name = "game-config"
              }
            }
          }
          
          env {
            name = "POD_NAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }
          
          ports {
            name           = "game"
            protocol       = "UDP"
            container_port = 7777
            policy         = "Dynamic"
            
            protection_protocol {
              value = "game-udp"
            }
          }
          
          config_files {
            name       = "server-cfg"
            mount_path = "/config/server.cfg"
          }
          
          volume_mounts {
            name       = "data"
            mount_path = "/data"
          }
        }
        
        health {
          disabled              = false
          initial_delay_seconds = 10
          period_seconds        = 5
          failure_threshold     = 3
        }
        
        termination_grace_period_seconds {
          value = 60
        }
        
        strategy {
          type = "RollingUpdate"
          
          rolling_update {
            max_surge {
              value = 1
            }
            max_unavailable {
              value = 0
            }
          }
        }
        
        volumes {
          name   = "data"
          medium = "Memory"
          size_limit {
            value = "1Gi"
          }
        }
        
        gateway_policies = ["ddos-protection"]
      }
    }
  }
}
```

**New (GameFabric) with Import:**
```terraform
# Import block for the existing armadaset
import {
  id = "prod/gameserver-set"
  to = gamefabric_armadaset.game_set
}

# Fetch the image ref
data "gamefabric_image" "gameserver" {
  branch = "prod"
  image  = "gameserver"
  tag    = "latest"  # or specific tag
}

# New resource definition
resource "gamefabric_armadaset" "game_set" {
  name        = "gameserver-set"
  environment = "prod"
  description = "Multi-region game server deployment"
  
  # Labels at armadaset level (previously metadata.labels)
  labels = {
    app     = "gameserver"
    managed = "terraform"
  }
  
  # Annotations at armadaset level (previously metadata.annotations)
  annotations = {
    owner = "platform-team"
  }
  
  # GameServer labels (previously spec.template.metadata.labels)
  gameserver_labels = {
    version = "v1.2.3"
    tier    = "production"
  }
  
  # GameServer annotations (previously spec.template.metadata.annotations)
  gameserver_annotations = {
    build = "12345"
  }
  
  # Autoscaling configuration
  autoscaling = {
    fixed_interval_seconds = 30
  }
  
  # Regions configuration (previously spec.armadas)
  # Each region can have its own replicas, labels, and envs
  regions = [
    {
      name = "europe"
      
      replicas = [
        {
          region_type  = "baremetal"
          min_replicas = 10
          max_replicas = 100
          buffer_size  = 10
        },
        {
          region_type  = "gcp"
          min_replicas = 5
          max_replicas = 50
          buffer_size  = 5
        }
      ]
      
      # Region-specific labels (override spec.override labels for this region)
      labels = {
        datacenter = "eu-central-1"
      }
      
      # Region-specific environment variables (override/extend template envs)
      envs = [
        {
          name  = "REGION"
          value = "europe"
        }
      ]
    },
    {
      name = "us-east"
      
      replicas = [
        {
          region_type  = "aws"
          min_replicas = 8
          max_replicas = 80
          buffer_size  = 8
        }
      ]
      
      # Region-specific labels
      labels = {
        datacenter = "us-east-1"
      }
      
      # Region-specific environment variables
      envs = [
        {
          name  = "REGION"
          value = "us-east"
        },
        {
          name  = "GAME_MODE"
          value = "ranked"  # Override the default "production" value
        }
      ]
    }
  ]
  
  # Template configuration (shared across all regions unless overridden)
  containers = [
    {
      name      = "default"
      image_ref = data.gamefabric_image.gameserver.image_ref
      
      resources = {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }
      
      envs = [
        {
          name  = "GAME_MODE"
          value = "production"
        },
        {
          name = "CONFIG_CONTENT"
          value_from = {
            config_file = "game-config"
          }
        },
        {
          name = "POD_NAME"
          value_from = {
            field_path = "metadata.name"
          }
        }
      ]
      
      ports = [
        {
          name                = "game"
          protocol            = "UDP"
          container_port      = 7777
          policy              = "Dynamic"
          protection_protocol = "game-udp"
        }
      ]
      
      config_files = [
        {
          name       = "server-cfg"
          mount_path = "/config/server.cfg"
        }
      ]
      
      volume_mounts = [
        {
          name       = "data"
          mount_path = "/data"
        }
      ]
    }
  ]
  
  # Health check configuration
  health_checks = {
    disabled              = false
    initial_delay_seconds = 10
    period_seconds        = 5
    failure_threshold     = 3
  }
  
  # Termination configuration
  termination_configuration = {
    grace_period_seconds = 60
  }
  
  # Rollout strategy
  strategy = {
    rolling_update = {
      max_surge       = 1
      max_unavailable = 0
    }
  }
  
  # Volumes
  volumes = [
    {
      name = "data"
      empty_dir = {
        size_limit = "1Gi"
      }
    }
  ]
  
  # Gateway policies
  gateway_policies = ["ddos-protection"]
  
  # Profiling (optional)
  profiling_enabled = false
}
```

**Key Changes:**
- **Flattened structure**: Removed `metadata` and `spec` wrappers throughout
- **Top-level attributes**:
  - `metadata.name` → `name`
  - `metadata.environment` → `environment`
  - `metadata.labels` → `labels` (armadaset-level labels)
  - `metadata.annotations` → `annotations` (armadaset-level annotations)
  - `spec.template.metadata.labels` → `gameserver_labels` (applied to game server pods)
  - `spec.template.metadata.annotations` → `gameserver_annotations` (applied to game server pods)
  - `spec.description` → `description`
- **Autoscaling**:
  - `spec.autoscaling.fixed_interval.seconds` → `autoscaling.fixed_interval_seconds` (now an object)
- **Armadas → Regions** (MAJOR restructuring):
  - `spec.armadas` → `regions` (renamed and restructured)
  - `armadas[].region` → `regions[].name` (field renamed)
  - `armadas[].description` field is no longer supported in the new provider (removed)
  - `armadas[].distribution` → `regions[].replicas` (nested list)
  - `distribution.name` → `replicas[].region_type`
  - Per-region overrides now embedded directly in each region object:
    - `spec.override[].labels` → `regions[].labels` (embedded in region)
    - `spec.override[].env` → `regions[].envs` (embedded in region)
- **Image references**:
  - `containers[].branch` and `containers[].image` → `containers[].image_ref` (use data source, same as Armadas)
- **Template changes** (same as regular armadas):
  - `spec.template.spec.containers` → `containers` (moved to top level)
  - `containers[].env` → `containers[].envs`
  - `containers[].env[].value_from.config_file_key_ref.name` → `containers[].envs[].value_from.config_file`
  - `containers[].env[].value_from.field_ref.field_path` → `containers[].envs[].value_from.field_path`
  - `containers[].ports[].protection_protocol.value` → `containers[].ports[].protection_protocol`
- **Health checks**: `spec.template.spec.health` → `health_checks`
- **Termination**: `spec.template.spec.termination_grace_period_seconds.value` → `termination_configuration.grace_period_seconds`
- **Strategy**: Removed `type` field, simplified `max_surge`/`max_unavailable` values
- **Volumes**: `volumes[].medium` and `volumes[].size_limit.value` → `volumes[].empty_dir.size_limit`
- **Gateway policies**: `spec.template.spec.gateway_policies` → `gateway_policies`
- **New attribute**: `profiling_enabled` (sets labels in the background)
- **Import ID format**: `"<environment>/<name>"`

**Important Notes:**
- The override mechanism changed completely: instead of a separate `spec.override` block, overrides are now embedded directly in each region configuration
- Region-specific `labels` and `envs` can be specified per region to override or extend the template values

### Migrating Gateway Policies

**Old (Nitrado EC):**
```terraform
resource "ec_protection_gatewaypolicy" "ddos_protection" {
  metadata {
    name = "ddos-protection"
  }
  
  spec {
    description       = "DDoS protection policy"
    destination_cidrs = ["0.0.0.0/0"]
  }
}
```

**New (GameFabric) with Import:**
```terraform
# Import block for the existing gateway policy
import {
  id = "ddos-protection"
  to = gamefabric_protection_gatewaypolicy.ddos_protection
}

# New resource definition
resource "gamefabric_protection_gatewaypolicy" "ddos_protection" {
  name              = "ddos-protection"
  description       = "DDoS protection policy"
  destination_cidrs = ["0.0.0.0/0"]
}
```

**Key Changes:**
- Removed `metadata` and `spec` wrappers
- `metadata.name` → `name`
- All spec attributes moved to top level
- Import ID format: `"<name>"`

### Migrating Image Updaters

**Image Updaters are a new resource in the GameFabric provider.** They were not available in the old Nitrado EC provider.

If you want to automatically update your armadas or armadasets when new images are pushed, you can create an ImageUpdater resource. This is what is created by the system if you select "autoUpdate" in the tag section in the frontend.

**Migration Notes:**
- There is no direct migration path as these didn't exist in the old provider
- If you want automatic image updates for your migrated armadas/armadasets, you'll need to create these resources from scratch
- For detailed examples and configuration options, see the [ImageUpdater resource documentation](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs/resources/imageupdater)
- No import is needed as these are new resources

## Step 4: Migration Workflow

Follow this workflow for each resource:

1. **Create the new resource definition and import block** using GameFabric provider syntax
2. **Run `terraform plan`** - Terraform will detect the import blocks and prepare to import
3. **Run `terraform apply`** to execute the import
4. **Verify no changes are detected** - Run `terraform plan` again to confirm
5. **If changes are detected**, adjust your configuration to match the existing resource
6. **Once plan shows no changes**, the migration for that resource is complete

**Note:** Do not remove old `ec_*` resources from your configuration yet. See Step 5 for proper cleanup procedures to avoid destroying your infrastructure.

### Example Migration Process

```bash
# 1. Add the new resource with import block to your .tf file
# (see examples above)

# 2. Review what will be imported and potential changes
terraform plan

# 3. Execute the import
terraform apply

# 4. Verify the configuration matches (should show no changes)
terraform plan

# 5. If you see unwanted changes, adjust your configuration
# and run plan again until you see "No changes"

# 6. Repeat for all resources before proceeding to Step 5 for cleanup
```

### Complete Example: Migrating an Environment

Here's a complete example showing the before and after state of migrating an environment:

**migration.tf (during migration):**
```terraform
# Import the existing environment into the new provider
import {
  id = "prod"
  to = gamefabric_environment.prod
}

# New GameFabric resource definition
resource "gamefabric_environment" "prod" {
  name         = "prod"
  display_name = "Production"
  description  = "Production environment"
  
  labels = {
    tier = "production"
  }
}
```

After running `terraform apply` and verifying with `terraform plan`, you can clean up:

**migration.tf (after successful migration - ready for Step 5):**
```terraform
# Import block can be removed in Step 5 after ALL resources are migrated
# Old ec_environment.prod will be removed from state in Step 5

# Keep only the new resource definition
resource "gamefabric_environment" "prod" {
  name         = "prod"
  display_name = "Production"
  description  = "Production environment"
  
  labels = {
    tier = "production"
  }
}
```

## Step 5: Complete Migration

Once **all** resources are migrated and showing no changes in `terraform plan`:

1. **Verify the new workspace** is working correctly:
   ```bash
   cd terraform-gamefabric
   terraform plan
   ```
   
   Should show no changes after all resources are imported.

2. **Test functionality** in the new workspace to ensure everything works as expected

3. **Keep the old workspace** as a backup:
   ```bash
   # Archive the old workspace
   cd ..
   tar -czf terraform-old-provider-backup-$(date +%Y%m%d).tar.gz terraform-old/
   ```
   
   **Do not delete the old workspace immediately!** Keep it for at least a few weeks to ensure the new setup is stable.

4. **Switch to using the new workspace** for all future changes

5. **Eventually decommission the old workspace** once you're confident the migration is successful (recommended: after 30+ days of stable operation)

## Troubleshooting

### Import Fails with "Resource Not Found"

Verify the resource exists and you're using the correct import ID format:

- Environments and Branches: `"<name>"`
- Environment-scoped resources: `"<environment>/<name>"`

### Plan Shows Unwanted Changes After Import

This usually means your Terraform configuration doesn't match the actual resource. Common causes:

- Missing or incorrect labels/annotations
- Different attribute formatting (e.g., CPU/memory values)
- Missing optional fields that have defaults

After importing, run `terraform show` to see the imported resource state and adjust your configuration to match.

### Authentication Issues

Ensure your service account has the same permissions in both providers. The authentication mechanism is the same, but verify:

- `GAMEFABRIC_PASSWORD` environment variable is set
- Service account has necessary permissions
- Customer ID is correct

### Replica Configuration Differences

The replica structure change from map to list is the most significant. Ensure you:

1. Convert the map structure to a list of objects
2. Add the `region_type` field to each replica object
3. Rename `min`/`max` to `min_replicas`/`max_replicas`
4. Maintain the same order if you care about priority (first in list has highest priority)

## Best Practices

1. **Migrate in stages**: Start with simpler resources (environments, branches, configfiles) before tackling complex ones (armadas, armadasets)
2. **Test in non-production first**: If possible, migrate a development environment first to validate your process
3. **Use version control**: Commit each successful resource migration separately for easy rollback
4. **Remove import blocks after success**: Import blocks are only needed once - remove them after successful import to keep your configuration clean
5. **Keep the old workspace as backup**: Don't delete your old Terraform workspace immediately - keep it for at least 30 days as a safety net
6. **Document custom changes**: Note any infrastructure-specific configurations or deviations from examples
7. **Verify functionality**: After migration, test that your game servers deploy and function correctly
8. **Plan for rollback**: Know how to switch back to the old workspace if critical issues arise

## Quick Reference: Import ID Formats

| Resource Type | Old Provider Name | Import ID Format | Example |
|---------------|-------------------|------------------|---------|
| Environment | `ec_core_environment` | `<name>` | `"prod"` |
| Branch | `ec_container_branch` | `<name>` | `"prod"` |
| ConfigFile | N/A (new resource) | `<environment>/<name>` | `"prod/game-config"` |
| Region | `ec_core_region` | `<environment>/<name>` | `"prod/europe"` |
| Armada | `ec_armada_armada` | `<environment>/<name>` | `"prod/gameserver"` |
| ArmadaSet | `ec_armada_armadaset` | `<environment>/<name>` | `"prod/gameserver-set"` |
| Gateway Policy | `ec_protection_gatewaypolicy` | `<name>` | `"ddos-protection"` |
| Image Updater | N/A (new resource) | `<environment>/<name>` | `"prod/auto-updater"` |

## Getting Help

If you encounter issues during migration:

- Check the [GameFabric documentation](https://docs.gamefabric.com)
- Review the [provider documentation](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs)
- Contact GameFabric support for infrastructure-specific questions

## Summary

The migration from Nitrado EC to GameFabric provider involves:

1. Creating a new Terraform workspace with the GameFabric provider
2. Writing new resource definitions using GameFabric syntax
3. Adding import blocks to import existing resources into the new Terraform state
4. Verifying configurations with `terraform plan`
5. Keeping the old workspace as a backup

The key changes across all resources are:

- **Flattened structure**: Removal of `metadata` and `spec` wrapper objects
- **Attribute renaming**: Consistent naming patterns (e.g., `env` → `envs`)
- **Label restructuring**: `metadata.labels` → `labels` for most resources; Armadas and ArmadaSets have more complex label mapping
- **Distribution → Replicas**: Armadas changed from repeated `distribution` blocks to `replicas` list with `region_type` field
- **Image references**: Containers now use `image_ref` (obtained from data source) instead of separate `branch` and `image` fields
- **New resources**: ConfigFiles and ImageUpdaters are now manageable Terraform resources (optional to use)
- **Import blocks**: Use declarative import blocks instead of imperative `terraform import` commands

By following this guide systematically and using a separate workspace with import blocks for each resource, you can migrate your infrastructure with zero downtime and minimal risk.
