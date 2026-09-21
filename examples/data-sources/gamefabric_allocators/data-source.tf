# Get all allocators without any filtering.
data "gamefabric_allocators" "all" {}

# Get allocators filtered by labels.
data "gamefabric_allocators" "eu" {
  label_filter = {
    region = "eu-west"
  }
}
