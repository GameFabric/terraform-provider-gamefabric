# Get an allocator by its unique name.
data "gamefabric_allocator" "this" {
  name = "my-allocator"
}

# Access the registry service endpoint to register or unregister new game servers:
# - data.gamefabric_allocator.this.registry.url
# - data.gamefabric_allocator.this.registry.active_token

# Access the allocation service endpoint to retrieve a game server:
# - data.gamefabric_allocator.this.allocation.url
# - data.gamefabric_allocator.this.allocation.active_token
