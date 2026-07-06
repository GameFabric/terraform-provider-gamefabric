# Get a ping discovery by its unique name.
data "gamefabric_ping_discovery" "this" {
  name = "my-ping-discovery"
}

# Replace "your_provider_your_resource" with your actual provider and resource type.
resource "your_provider_your_resource" "this" {
  ping_discovery_url   = data.gamefabric_ping_discovery.this.url
  ping_discovery_token = data.gamefabric_ping_discovery.this.active_token
}
