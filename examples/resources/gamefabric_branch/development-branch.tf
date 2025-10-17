resource "gamefabric_branch" "feature" {
  name         = "feature"
  display_name = "Feature Branch"
  description  = "Feature development branch with image retention policies"

  labels = {
    environment = "development"
  }

  # Define retention policy rules for managing image lifecycle.
  # Multiple rules can be specified to handle different image patterns.
  retention_policy_rules = [
    {
      name        = "keep-latest-releases" # Required: Name of the retention policy rule.
      image_regex = ".*"                   # Optional: Regex to match image names. Empty or ".*" matches all.
      tag_regex   = ".*"                   # Optional: Regex to match tag names (e.g., semantic versions).
      keep_count  = 5                      # Optional: Minimum number of matching tags to retain per image.
      keep_days   = 1                      # Optional: Minimum number of days to keep matching tags.
    },
  ]
}
