resource "gamefabric_branch" "prod" {
  name         = "prod"
  display_name = "Production Branch"
  description  = "Production branch for stable releases"

  labels = {
    environment = "production"
    team        = "platform"
    managed_by  = "terraform"
  }
  retention_policy_rules = [
    {
      name        = "keep-latest-releases" # Required: Name of the retention policy rule.
      image_regex = ".*"                   # Optional: Regex to match image names. Empty or ".*" matches all.
      tag_regex   = ".*"                   # Optional: Regex to match tag names (e.g., semantic versions).
      keep_count  = 10                     # Optional: Minimum number of matching tags to retain per image.
      keep_days   = 14                     # Optional: Minimum number of days to keep matching tags.
    },
  ]
}
