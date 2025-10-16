# Create a simple branch with minimal configuration.
# The name is optional and will be auto-generated if not provided.
# A branch represents a container image repository branch in the GameFabric system.
resource "gamefabric_branch" "main" {
  name         = "main"                # Optional: Unique identifier. If omitted, will be auto-generated.
  display_name = "Main Branch"         # Required: User-friendly display name for the branch.
  description  = "Main production branch for container images"
}

# Create a branch with labels for organization.
# Labels are key-value pairs that help categorize and filter branches.
resource "gamefabric_branch" "develop" {
  name         = "develop"
  display_name = "Development Branch"
  description  = "Development and testing branch"
  
  labels = {
    environment = "development"
    team        = "platform"
    managed_by  = "terraform"
  }
}

# Create a branch with retention policies.
# Retention policies control how long container images and tags are kept.
resource "gamefabric_branch" "feature" {
  name         = "feature"
  display_name = "Feature Branch"
  description  = "Feature development branch with image retention policies"

  labels = {
    environment = "development"
    branch_type = "feature"
  }

  # Define retention policy rules for managing image lifecycle.
  # Multiple rules can be specified to handle different image patterns.
  retention_policy_rules = [
    {
      name       = "keep-latest-releases"          # Required: Name of the retention policy rule.
      image_regex = ".*"                            # Optional: Regex to match image names. Empty or ".*" matches all.
      tag_regex   = "^v[0-9]+\\.[0-9]+\\.[0-9]+$"  # Optional: Regex to match tag names (e.g., semantic versions).
      keep_count  = 10                              # Optional: Minimum number of matching tags to retain per image.
      keep_days   = 90                              # Optional: Minimum number of days to keep matching tags.
    },
    {
      name       = "keep-snapshot-builds"
      image_regex = ".*"
      tag_regex   = "^snapshot-.*"
      keep_count  = 5   # Keep at least 5 snapshot builds
      keep_days   = 30  # Keep snapshots for at least 30 days
    },
    {
      name       = "delete-old-pr-builds"
      image_regex = ".*"
      tag_regex   = "^pr-[0-9]+-.*"
      keep_count  = 3   # Only keep last 3 PR builds per image
      keep_days   = 7   # Delete PR builds older than 7 days
    }
  ]
}

# Create a branch with comprehensive retention policies.
# This example shows more complex retention rules for different image patterns.
resource "gamefabric_branch" "production" {
  display_name = "Production Branch"
  description  = "Production branch with strict retention policies"

  labels = {
    environment = "production"
    criticality = "high"
  }

  retention_policy_rules = [
    {
      name       = "production-releases"
      image_regex = "^app-.*"               # Match images starting with 'app-'
      tag_regex   = "^release-.*"           # Match tags starting with 'release-'
      keep_count  = 20                      # Keep last 20 production releases
      keep_days   = 365                     # Keep for at least 1 year
    },
    {
      name       = "stable-versions"
      image_regex = ".*"                    # Match all images
      tag_regex   = "^stable$|^latest$"     # Match 'stable' or 'latest' tags
      keep_count  = 100                     # Keep many stable versions
      keep_days   = 730                     # Keep for 2 years
    }
  ]
}

# Output the branch information for reference.
output "main_branch_id" {
  description = "The ID of the main branch"
  value       = gamefabric_branch.main.id
}

output "feature_branch_retention_rules" {
  description = "The retention policy rules applied to the feature branch"
  value       = gamefabric_branch.feature.retention_policy_rules
}
