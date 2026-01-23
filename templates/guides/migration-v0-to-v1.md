---
page_title: "Migrating to GameFabric Provider v1.0"
subcategory: "Guides"
description: |-
  Learn how to upgrade Terraform configuration from GameFabric Provider v0.x to v1.0.
---

# Migrating to GameFabric Provider v1.0

This guide helps upgrade Terraform configuration from GameFabric Provider v0.x to v1.0.

> **Note:** This guide applies when upgrading from any v0.x version (v0.3, v0.7, etc.) to v1.0 or later. Skip this guide if already on v1.0.

⚠️ **CRITICAL REQUIREMENT:** This migration only works on GameFabric Provider v0.7 or later. If running v0.3 to v0.6, first upgrade to v0.7 using standard Terraform upgrade procedures, then return to follow this guide.

## What's Changing in v1.0

GameFabric Provider v1.0 removes the deprecated `password` attribute from the `gamefabric_service_account` resource. 

**Why this change?**
- In v0.7: `gamefabric_service_account_password` resource was introduced to separate password management from service account management
- In v1.0: The old `password` attribute on `gamefabric_service_account` is removed because the new resource handles password management

The `gamefabric_service_account_password` resource is available in v0.7+ and provides a better way to manage service account passwords separately.

Migrate to the new `gamefabric_service_account_password` resource if currently using service account passwords in Terraform configuration.

## Prerequisites

**Important:** This migration requires a two-phase upgrade process because the new password management resource was added in v0.7.

**Timeline of changes:**
- **v0.3 through v0.6:** Only `gamefabric_service_account` with inline `password` attribute exists
- **v0.7:** `gamefabric_service_account_password` resource introduced to separate password management
- **v1.0:** `password` attribute removed from `gamefabric_service_account`; only `gamefabric_service_account_password` is supported

**Why two phases are needed:**

The migration cannot happen in one step because:
1. Versions before v0.7 don't have the `gamefabric_service_account_password` resource
2. Cannot remove the `password` attribute (v1.0) without having the new resource first
3. Must have both resources available during migration to replace the old approach with the new one

**Two-phase migration process:**

**Phase 1: Upgrade to v0.7 (if not already on v0.7+)**
1. If currently on v0.3, v0.4, v0.5, or v0.6: Upgrade to v0.7 first
2. This adds the `gamefabric_service_account_password` resource
3. Do NOT follow this migration guide yet
4. Wait until v0.7 is stable in the environment

**Phase 2: Migrate to v1.0 (this guide)**
1. Once confirmed running v0.7 or later
2. Follow the 8 migration steps in this guide
3. Make the password management changes while on v0.7
4. Then upgrade to v1.0

**Cannot upgrade directly:** Direct upgrades from v0.3-v0.6 to v1.0 are not supported because:
- Versions v0.3-v0.6 don't have the new `gamefabric_service_account_password` resource
- The `password` attribute cannot be replaced without the new resource existing first
- Migration must use the new resource introduced in v0.7

## Do I Need to Migrate?

**Migrate if:**
- Terraform configuration includes `gamefabric_service_account` resources that reference the `password` attribute

**Skip this guide if:**
- Terraform configuration doesn't use service accounts
- Already using `gamefabric_service_account_password` resource

## Migration Steps

### Overview: Complete Upgrade Workflow

This is the complete workflow for upgrading from any v0.x version to v1.0:

**For users on v0.3 - v0.6:**
```
v0.3/v0.4/v0.5/v0.6 → [Upgrade] → v0.7 → [Follow this guide] → v1.0
                        (Step 1)                (Steps 2-8)
```

**For users already on v0.7+:**
```
v0.7 → [Follow this guide] → v1.0
            (Steps 1-8)
```

The following 8 steps assume already on v0.7 or later. If on earlier versions, first upgrade to v0.7 using normal Terraform upgrade procedures, then return to Step 1.

### Step 1: Check current configuration

Look for `gamefabric_service_account` resources in `.tf` files that reference the `password` attribute.

**Example of old configuration that needs migration:**

```terraform
resource "gamefabric_service_account" "ci_bot" {
  name = "ci-bot"
  labels = {
    team = "platform"
  }
}

# Using the password somewhere
output "ci_bot_password" {
  value     = gamefabric_service_account.ci_bot.password
  sensitive = true
}
```

### Step 2: Add the new password resource

Add a `gamefabric_service_account_password` resource for each service account where the password is needed.

**New configuration:**

```terraform
# Keep existing service account resource
resource "gamefabric_service_account" "ci_bot" {
  name = "ci-bot"
  labels = {
    team = "platform"
  }
}

# Add the new password resource
resource "gamefabric_service_account_password" "ci_bot" {
  service_account = gamefabric_service_account.ci_bot.name
}

# Update output to use the new password resource
output "ci_bot_password" {
  value     = gamefabric_service_account_password.ci_bot.password
  sensitive = true
}
```

### Step 3: Update all password references

Find and replace all references to the old password attribute with the new resource.

**Before:**
```terraform
gamefabric_service_account.ci_bot.password
```

**After:**
```terraform
gamefabric_service_account_password.ci_bot.password
```

### Step 4: Run Terraform plan

Before upgrading to v1.0, test changes with current provider version:

```bash
terraform plan
```

Terraform displays:
- New `gamefabric_service_account_password` resources to be created
- Existing `gamefabric_service_account` resources remain unchanged

### Step 5: Apply the changes

Apply the changes while still on v0.x:

```bash
terraform apply
```
⚠️ **Important:** This operation resets the service account password. The new `gamefabric_service_account_password` resource generates a fresh password.
### Step 6: Save the new password

After applying, retrieve and save the new password:

```bash
terraform output ci_bot_password
```

Store this password securely (e.g., in a secret management system, CI/CD platform, etc.).

### Step 7: Update systems using the password

Update all systems that use the service account password with the new password value:
- CI/CD pipelines
- Application configurations
- Scripts and automation tools
- Secret management systems

### Step 8: Update to v1.0

Now safely upgrade to v1.0. Update provider version in Terraform configuration:

```terraform
terraform {
  required_providers {
    gamefabric = {
      source  = "gamefabric/gamefabric"
      version = "~> 1.0"  # Update from ~> 0.x
    }
  }
}
```

Run:

```bash
terraform init -upgrade
terraform plan
```

The plan shows no changes because migration to the new password resource has already been completed.

## Complete example

Here's a complete before and after example:

### Before (v0.x)

```terraform
terraform {
  required_providers {
    gamefabric = {
      source  = "gamefabric/gamefabric"
      version = "~> 0.7"
    }
  }
}

resource "gamefabric_service_account" "ci_bot" {
  name = "ci-bot"
  labels = {
    team    = "platform"
    purpose = "ci-cd"
  }
}

resource "gamefabric_service_account" "monitoring" {
  name = "monitoring-agent"
  labels = {
    team = "ops"
  }
}

# Using passwords from service accounts
output "ci_password" {
  value     = gamefabric_service_account.ci_bot.password
  sensitive = true
}

output "monitoring_password" {
  value     = gamefabric_service_account.monitoring.password
  sensitive = true
}
```

### After (v1.0)

```terraform
terraform {
  required_providers {
    gamefabric = {
      source  = "gamefabric/gamefabric"
      version = "~> 1.0"
    }
  }
}

# Service account resources stay the same
resource "gamefabric_service_account" "ci_bot" {
  name = "ci-bot"
  labels = {
    team    = "platform"
    purpose = "ci-cd"
  }
}

resource "gamefabric_service_account" "monitoring" {
  name = "monitoring-agent"
  labels = {
    team = "ops"
  }
}

# New password resources
resource "gamefabric_service_account_password" "ci_bot" {
  service_account = gamefabric_service_account.ci_bot.name
}

resource "gamefabric_service_account_password" "monitoring" {
  service_account = gamefabric_service_account.monitoring.name
}

# Updated outputs
output "ci_password" {
  value     = gamefabric_service_account_password.ci_bot.password
  sensitive = true
}

output "monitoring_password" {
  value     = gamefabric_service_account_password.monitoring.password
  sensitive = true
}
```

## Important Notes

### Password will be reset

⚠️ **Warning:** Creating the `gamefabric_service_account_password` resource generates a new password. The old password no longer works.

**Required actions:**
1. Apply the migration changes
2. Retrieve the new password using `terraform output`
3. Update the password in all systems that use it (CI/CD, applications, scripts, etc.)
4. Test that everything works with the new password before upgrading to v1.0

### State management

Store the Terraform state file securely:
- Encrypt data at rest using remote backends (Terraform Cloud, S3 with encryption, etc.)
- Restrict access to authorized team members only
- Never commit state files to version control

For more information, see [Terraform's documentation on managing sensitive data](https://developer.hashicorp.com/terraform/language/manage-sensitive-data).

### Why this change?

Separating password management into its own resource provides:
- **Better security:** Manage passwords independently without recreating the service account
- **More flexibility:** Rotate passwords by updating the password resource
- **Clearer intent:** Explicitly declare when password access is needed versus managing the account

## Getting help

If issues occur during migration:

- Check the [GameFabric Provider documentation](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs)
- Review the [service_account resource docs](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs/resources/service_account)
- Review the [service_account_password resource docs](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs/resources/service_account_password)
- Contact the Customer Success Management team for assistance

## Summary

**Important:** This migration requires two phases and only works on v0.7+.

**Phase 1 (if needed): Upgrade to v0.7**
- If on v0.3 to v0.6: First upgrade to v0.7 using standard Terraform procedures
- Skip to Phase 2 if already on v0.7 or later

**Phase 2: Migrate from v0.7 to v1.0 (this guide)**

To migrate the password attribute from v0.7 to v1.0:

1. **Ensure** running GameFabric Provider v0.7 or later
2. **Identify** service accounts that reference the `password` attribute
3. **Create** new `gamefabric_service_account_password` resources
4. **Update** all password references to use the new resource
5. **Apply** changes while still on v0.7 (this resets passwords)
6. **Retrieve** and save the new passwords
7. **Update** all systems using the service account passwords
8. **Upgrade** to v1.0 provider version

Following this two-phase process creates a smooth migration with minimal downtime. Passwords reset during migration, so plan accordingly and update all dependent systems.
