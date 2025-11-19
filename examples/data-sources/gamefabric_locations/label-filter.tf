data "gamefabric_locations" "us_east" {
  label_filter = {
    "g8c.io/country" = "us"
    "g8c.io/state" = [
      "sc", # South Carolina
      "oh", # Ohio
      "fl", # Florida
      "ny", # New York
      "dc", # District of Columbia
    ]
  }
}

data "gamefabric_locations" "eu_west" {
  continent = "europe"
  label_filter = {
    "g8c.io/city" = ["amsterdam", "frankfurt", "london"]
  }
}
