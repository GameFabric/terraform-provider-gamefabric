resource gamefabric_group "my_group" {
  name = "my-group"
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
