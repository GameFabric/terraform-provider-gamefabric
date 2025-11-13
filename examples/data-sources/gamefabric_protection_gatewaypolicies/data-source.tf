# Get all gateway policies without any filtering.
data "gamefabric_protection_gatewaypolicies" "all" {
}

# Get gateway policies filtered by labels.
data "gamefabric_protection_gatewaypolicies" "by_labels" {
  label_filter = {
    stage = "dev"
  }
}
