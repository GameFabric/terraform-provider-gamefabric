resource gamefabric_group "developer_group" {
  name = "developer-group"
  labels = {
    team = "backend"
  }
  annotations = {
    used-by = "ci"
  }
  users = [
    "gamedev@example.com"
  ]
}
