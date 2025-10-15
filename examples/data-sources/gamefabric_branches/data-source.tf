# Get all branches
data "gamefabric_branches" "all" {
}

# Get branches by labels (exact matches of all provided labels)
data "gamefabric_branches" "baremetal" {
  label_filter = {
    baremetal = "true"
  }
}
