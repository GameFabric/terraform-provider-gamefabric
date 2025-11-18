data "gamefabric_locations" "us-east" {
  label_filter = {
    "g8c.io/country" = "us"
    "g8c.io/state" = [
      "sc", # South Carolina, us-east1
      "oh", # Ohio, us-east5
    ]
  }
}

data "gamefabric_locations" "eu-west" {
  continent = "europe"
  label_filter = {
    "g8c.io/city" = ["amsterdam", "frankfurt", "london"]
  }
}
