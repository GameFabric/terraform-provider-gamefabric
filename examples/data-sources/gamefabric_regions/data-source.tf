# Get all regions without any filtering.
data "gamefabric_regions" "all" {
  environment = "dev" # The environment this region belongs to (required)
}

# Get regions filtered by labels.
#
# Labels are key-value pairs used to organize and categorize regions.
#
# Note:
# - Only regions that have ALL specified labels with exact matching values are returned.
data "gamefabric_regions" "eu" {
  environment = "dev" # The environment this region belongs to (required)
  label_filter = {
    env = "prod"
  }
}
# The returned output contains the resource limits for each region type.
# output:
# [
#   {
#     name = ""
#     environment = ""
#     types = {
#       baremetal = {
#         cpu = "4"
#         memory = "8Gi"
#       }
#     }
#   },
#   {
#     name = ""
#     environment = ""
#     types = {
#       baremetal = {
#         cpu = "4"
#         memory = "8Gi"
#       }
#     }
#   }
# ]
