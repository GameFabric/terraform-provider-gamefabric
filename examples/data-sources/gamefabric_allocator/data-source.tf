# Get an allocator by its unique name.
data "gamefabric_allocator" "this" {
  name = "my-allocator"
}

# Access the registry service endpoint to register or unregister new game servers:

output "registry_url" {
  value = data.gamefabric_allocator.this.registry_url
}

output "registry_active_token" {
  value     = data.gamefabric_allocator.this.registry_active_token
  sensitive = true
}

# Access the allocation service endpoint to retrieve a registered game server:

output "allocation_url" {
  value = data.gamefabric_allocator.this.allocation_url
}

output "allocation_active_token" {
  value     = data.gamefabric_allocator.this.allocation_active_token
  sensitive = true
}
