# Create a simple branch with minimal configuration.
resource "gamefabric_branch" "main" {
  name         = "main"                # Required: Unique identifier.
  display_name = "Main Branch"         # Required: User-friendly display name for the branch.
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
  retention_policy_rules = []
}

# Create a branch with retention policies.
# Retention policies control how long container images and tags are kept.
# If keep_count and keep_days are both specified, the policy retains images that meet either condition.
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
      name        = "keep-latest-releases"         # Required: Name of the retention policy rule.
      image_regex = ".*"                           # Optional: Regex to match image names. Empty or ".*" matches all.
      tag_regex   = "^v[0-9]+\\.[0-9]+\\.[0-9]+$"  # Optional: Regex to match tag names (e.g., semantic versions).
      keep_count  = 10                             # Optional: Minimum number of matching tags to retain per image.
      keep_days   = 90                             # Optional: Minimum number of days to keep matching tags.
    },
    {
      name       = "keep-snapshot-builds"
      image_regex = ".*"
      tag_regex   = "^snapshot-.*"
      keep_count  = 5   # Keep at least 5 snapshot builds, and
      keep_days   = 30  # Keep snapshots for at least 30 days
    },
    {
      name       = "delete-old-pr-builds"
      image_regex = ".*"
      tag_regex   = "^pr-[0-9]+-.*"
      keep_count  = 3   # Only keep last 3 PR builds per image,
      keep_days   = 7   # Delete PR builds older than 7 days
    }
  ]
}
