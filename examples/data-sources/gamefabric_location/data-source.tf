# Get a specific location by its unique name.
data "gamefabric_location" "frankfurt" {
  name = "eu-central-1-fra"
}
# output: .names = ["frankfurt-gcp", "amsterdam-gcp"] # this is a list of locations that match the criteria

# Use the location data source to validate availability.
# This is useful when you need to ensure a location is available
# before using it in region or other resource configurations.
data "gamefabric_location" "us_east" {
  name = "us-east-1-nyc"
}
