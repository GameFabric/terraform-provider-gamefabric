# Create a token service using custom platform signing keys.
resource "gamefabric_steelshield_tokenservice" "custom_keys" {
  name             = "my-token-service"
  development_mode = false

  labels = {
    team = "platforms"
  }

  game_name = "my-game"

  platforms = {
    pc = {
      game_client_token_keys = [
        {
          key               = "-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMttWYwVnWHzIKQ7wdY7jtcTRMseUloX\nbWexrZWBQnNkmRq+Cn3YUyLT6+TpCFmlcV5Zga737wjhnXlMA7F616ECAwEAAQ==\n-----END PUBLIC KEY-----\n"
          signing_algorithm = "RS256"
        }
      ]
    }
    xbox = {
      game_client_token_keys = [
        {
          key               = "-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMttWYwVnWHzIKQ7wdY7jtcTRMseUloX\nbWexrZWBQnNkmRq+Cn3YUyLT6+TpCFmlcV5Zga737wjhnXlMA7F616ECAwEAAQ==\n-----END PUBLIC KEY-----\n"
          signing_algorithm = "RS256"
        }
      ]
    }
  }
}

# Create a token service for Epic Online Services (EOS).
resource "gamefabric_steelshield_tokenservice" "eos" {
  name             = "my-eos-service"
  development_mode = false

  game_name = "my-game"

  eos = [
    {
      client_id     = "1234567890abcdef1234567890abcdef"
      product_id    = "abc123"
      deployment_id = "def456"
      sandbox_id    = "ghi789"
      # token_types defaults to ["connect"].
    },
  ]
}

# Create a token service backed by a JWKS endpoint.
resource "gamefabric_steelshield_tokenservice" "jwks" {
  name             = "my-jwks-service"
  development_mode = false

  game_name = "my-game"

  platforms = {
    pc = {}
  }

  jwks = {
    url = "https://tokens.mygame.com/.well-known/jwks.json"
  }
}
