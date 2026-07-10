# Get a ping discovery by its unique name.
data "gamefabric_ping_discovery" "this" {
  name = "my-ping-discovery"
}

# Access the endpoint to retrieve ping targets:

output "ping_discovery_url" {
  value = data.gamefabric_ping_discovery.this.url
}

output "ping_discovery_active_token" {
  value     = data.gamefabric_ping_discovery.this.active_token
  sensitive = true
}
