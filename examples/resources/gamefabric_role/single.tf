resource "gamefabric_role" "developer_role" {
  name = "developer-role"

  labels = {
    team = "backend"
    env  = "development"
  }

  annotations = {
    description = "Role for developers with limited read and write access"
  }

  rules = [
    {
      api_groups     = ["armada"]
      environments   = ["dev"]
      resources      = ["armadas", "armadagameserverstates", "armadarevisions", "armadarevisionstatuses", "armadasets", "armadasetrevisions"],
      verbs          = ["get"]
      resource_names = ["development-armada"]
    },
    {
      api_groups   = ["core"]
      environments = ["dev"]
      resources    = ["regions"]
      verbs        = ["get", "post"]
    },
    {
      api_groups   = ["core"]
      environments = ["*"]
      resources    = ["sites", "locations"]
      verbs        = ["get"]
    },
    {
      api_groups   = ["container"]
      environments = ["*"]
      resources    = ["images"]
      verbs        = ["get"]
      scopes       = ["dev-branch"]
    },
    {
      api_groups   = ["observability"]
      environments = ["*"]
      resources    = ["monitoring"]
      verbs        = ["get", "post"]
    },
    {
      api_groups   = ["audit"]
      environments = ["*"]
      resources    = ["logs"]
      verbs        = ["get"]
    },
    {
      api_groups   = ["core"]
      environments = ["*"]
      resources    = ["sites/gameservers"]
      verbs        = ["delete"]
    },
  ]
}
