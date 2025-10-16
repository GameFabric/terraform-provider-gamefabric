# Get all regions without any filtering.
data "gamefabric_regions" "all" {}

# Get regions filtered by labels.
#
# Labels are key-value pairs used to organize and categorize environments.
#
# Note:
# - Only environments that have ALL specified labels with exact matching values are returned.
data "gamefabric_regions" "eu" {
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
