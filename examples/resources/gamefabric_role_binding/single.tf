resource "gamefabric_role_binding" "developer_binding" {
  role = "developer-role"
  groups = [
    "developer-group",
  ]
  users = [
    "ci-workflow@ec.nitrado.systems"
  ]
}
