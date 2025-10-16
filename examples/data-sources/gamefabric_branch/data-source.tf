# Get a branch by its unique name.
# The name must match exactly and is case-sensitive.
data "gamefabric_branch" "main" {
  name = "main"
}

# Alternatively, get a branch by display name.
# Display names are user-friendly names that can contain spaces and special characters.
# Note: Only one of 'name' or 'display_name' should be specified.
data "gamefabric_branch" "production" {
  display_name = "Production Branch"
}

# Access branch attributes after retrieval.
# The branch data source returns information about retention policies and other metadata.
output "branch_retention_rules" {
  description = "The retention policy rules for the branch"
  value       = data.gamefabric_branch.main.retention_policy_rules
}
