---
page_title: "Migrating to GameFabric Provider v1.0"
subcategory: "Guides"
description: |-
  Learn how to upgrade your Terraform configuration from GameFabric Provider v0.x to v1.0.
---

# Migrating to GameFabric Provider v1.0

This guide helps you upgrade your Terraform configuration from GameFabric Provider v0.x to v1.0.

## What's Changing in v1.0

GameFabric Provider v1.0 removes the deprecated `password` attribute from the `gamefabric_service_account` resource. If you're currently using service account passwords in your Terraform configuration, you'll need to migrate to the new `gamefabric_service_account_password` resource.

## Do I Need to Migrate?

**You need to migrate if:**
- You have `gamefabric_service_account` resources in your Terraform configuration
- You're currently reading or using the `password` attribute from these resources

**You don't need to migrate if:**
- You're not using service accounts in Terraform
- You're already using the `gamefabric_service_account_password` resource
- You don't access the password attribute at all

## Migration Steps

### Step 1: Check Your Current Configuration

Look for any `gamefabric_service_account` resources in your `.tf` files that reference the `password` attribute.

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

### Step 2: Add the New Password Resource

Add a `gamefabric_service_account_password` resource for each service account where you need the password.

**New configuration:**

```terraform
# Keep your existing service account resource
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

### Step 3: Update All Password References

Find and replace all references to the old password attribute with the new resource.

**Before:**
```terraform
gamefabric_service_account.ci_bot.password
```

**After:**
```terraform
gamefabric_service_account_password.ci_bot.password
```

### Step 4: Run Terraform Plan

Before upgrading to v1.0, test your changes with your current provider version:

```bash
terraform plan
```

You should see that Terraform will:
- Create new `gamefabric_service_account_password` resources
- The existing `gamefabric_service_account` resources remain unchanged

### Step 5: Apply the Changes

Apply the changes while still on v0.x:

```bash
terraform apply
```

**Important:** This will reset the service account password! The new `gamefabric_service_account_password` resource will generate a fresh password.

### Step 6: Save the New Password

After applying, retrieve and save the new password:

```bash
terraform output ci_bot_password
```

Store this password securely (e.g., in your secret management system, CI/CD platform, etc.).

### Step 7: Update Systems Using the Password

Update all systems that use the service account password with the new password value:
- CI/CD pipelines
- Application configurations
- Scripts and automation tools
- Secret management systems

### Step 8: Update to v1.0

Now you can safely upgrade to v1.0. Update your provider version in your Terraform configuration:

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

The plan should show no changes because you've already migrated to the new password resource.

## Complete Example

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

### Password Will Be Reset

⚠️ **Warning:** When you create the `gamefabric_service_account_password` resource, a new password will be generated. The old password will no longer work.

**What you need to do:**
1. Apply the migration changes
2. Retrieve the new password using `terraform output`
3. Update the password in all systems that use it (CI/CD, applications, scripts, etc.)
4. Test that everything works with the new password before upgrading to v1.0

### State Management

The password is stored in your Terraform state file as a sensitive value. Make sure your state file is:
- Encrypted at rest (use remote backends like Terraform Cloud, S3 with encryption, etc.)
- Access-controlled (only authorized team members can access it)
- Not committed to version control

For more information, see [Terraform's documentation on managing sensitive data](https://developer.hashicorp.com/terraform/language/manage-sensitive-data).

### Why This Change?

Separating password management into its own resource provides:
- **Better security:** You can manage passwords independently without recreating the service account
- **More flexibility:** You can rotate passwords by updating the password resource
- **Clearer intent:** It's explicit when you need password access vs. just managing the account

## Troubleshooting

### "Resource not found" error after migration

If you see errors about the service account not being found, make sure you've applied the changes before upgrading to v1.0.

### Password stopped working

This is expected after migration. The `gamefabric_service_account_password` resource generates a new password. Retrieve it with:

```bash
terraform output -raw ci_bot_password
```

Then update all systems using the old password.

### Want to keep the old password?

Unfortunately, Terraform doesn't have access to the existing password after it's been created. You'll need to accept the new password and update your systems accordingly.

### Multiple service accounts to migrate

If you have many service accounts, consider migrating them one at a time or in small batches to minimize disruption:

1. Create the password resource for one service account
2. Apply and retrieve the new password
3. Update all systems using that service account
4. Verify everything works
5. Repeat for the next service account

## Getting Help

If you encounter issues during migration:

- Check the [GameFabric Provider documentation](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs)
- Review the [service_account resource docs](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs/resources/service_account)
- Review the [service_account_password resource docs](https://registry.terraform.io/providers/gamefabric/gamefabric/latest/docs/resources/service_account_password)
- Contact your Customer Success Management team for assistance

## Summary

To migrate from GameFabric Provider v0.x to v1.0:

1. **Identify** service accounts that reference the `password` attribute
2. **Create** new `gamefabric_service_account_password` resources
3. **Update** all password references to use the new resource
4. **Apply** changes while still on v0.x (this resets passwords!)
5. **Retrieve** and save the new passwords
6. **Update** all systems using the service account passwords
7. **Upgrade** to v1.0 provider version

By following these steps, you'll have a smooth migration with minimal downtime. Just remember that passwords will be reset, so plan accordingly and update all dependent systems.
