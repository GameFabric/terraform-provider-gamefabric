# Get all branches without filtering.
# This returns all container image branches in the system.
# Branches represent container image repositories or streams.
data "gamefabric_branches" "all" {
}

# Get branches filtered by labels.
# Labels are key-value pairs used to organize and categorize branches.
# Only branches that have ALL specified labels with exact matching values are returned.
data "gamefabric_branches" "baremetal" {
  label_filter = {
    baremetal = "true"
  }
}

# Get branches with multiple label filters.
# This example filters for production branches managed by a specific team.
data "gamefabric_branches" "production_platform" {
  label_filter = {
    environment = "production"
    team        = "platform"
  }
}

# Get development branches.
data "gamefabric_branches" "development" {
  label_filter = {
    environment = "development"
  }
}

# Access the list of branches returned by the data source.
# Each branch includes name, display_name, labels, description, and retention_policy_rules.
output "all_branch_names" {
  description = "Names of all branches"
  value       = [for branch in data.gamefabric_branches.all.branches : branch.name]
}

output "production_branches_with_retention" {
  description = "Production branches and their retention policies"
  value = {
    for branch in data.gamefabric_branches.production_platform.branches :
    branch.name => {
      display_name         = branch.display_name
      retention_rule_count = length(branch.retention_policy_rules)
    }
  }
}
