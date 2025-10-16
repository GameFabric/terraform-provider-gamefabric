# Get all branches without filtering.
data "gamefabric_branches" "all" {
}

# Get branches filtered by labels.
#
# Labels are key-value pairs used to organize and categorize branches.
# Only branches that have ALL specified labels with exact matching values are returned.
data "gamefabric_branches" "production" {
  label_filter = {
    env = "prod"
  }
}
# output
# [
#   {
#     name = "production"
#     retention_policy_rules = [
#       {
#         name        = "default"
#         keep_count  = 10
#         keep_days   = 30
#         image_regex = ".*" # this is a regex to match the image name, it should be used to match the image name
#         tag_regex   = ".*" # this is a regex to match the tag, it should be used to match the tag
#       }
#     ]
#   }
# ]

# Get branches with multiple label filters.
data "gamefabric_branches" "production_platform" {
  label_filter = {
    env  = "prod"
    team = "platform"
  }
}
