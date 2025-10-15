# Get all branches
data "gamefabric_branches" "all" {
}

# Get branches by name (exact match)
data "gamefabric_branches" "by_name" {
  name = "test"
}

# Get branches by display name (exact match)
data "gamefabric_branches" "by_display_name" {
  display_name = "Test Branch"
}

# Get branches by labels (exact matches of all provided labels)
data "gamefabric_branches" "baremetal" {
  label_filter = {
    baremetal = "true"
  }
}

# Combined filters
data "gamefabric_branches" "filtered" {
  display_name = "Test Branch"
  label_filter = {
    baremetal = "true"
  }
}
