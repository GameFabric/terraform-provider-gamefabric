resource "gamefabric_service_account" "ci_workflow" {
  name = "ci-workflow"
  labels = {
    team = "backend"
    env  = "development"
  }
}
