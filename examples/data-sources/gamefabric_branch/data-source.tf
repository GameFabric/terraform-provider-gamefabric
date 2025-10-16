# Get a branch by its unique name.
#
# The name must match exactly and is case-sensitive.
data "gamefabric_branch" "main" {
  name = "main"
}
# output:
#   {
#     name = "main"
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

# Get a branch by display name.
#
# Display names are user-friendly names that can contain spaces and special characters.
#
# Note:
# - Only one of 'name' or 'display_name' should be specified.
# - If multiple branches share the same display name, an error will occur.
data "gamefabric_branch" "production" {
  display_name = "Production Branch"
}