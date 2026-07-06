# Get a ping discovery by its unique name.
data "gamefabric_ping_discovery" "this" {
  name = "my-ping-discovery"
}

# Access the datasource values for further usage in other resources:
# - data.gamefabric_ping_discovery.this.url
# - data.gamefabric_ping_discovery.this.active_token
