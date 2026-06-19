resource "gamefabric_cloudbudget" "percentage" {
  name       = "percentage"
  max_budget = 10000
  receivers  = [gamefabric_notification_receiver.accounting.name]
  thresholds = [
    "50%",
    "90%",
  ]
}

resource "gamefabric_cloudbudget" "absolute" {
  name       = "absolute"
  max_budget = 10000
  receivers  = [gamefabric_notification_receiver.accounting.name]
  thresholds = [
    "5000",
    "9000",
  ]
}

resource "gamefabric_cloudbudget" "every_10k_dollar" {
  name       = "every-10k-dollar"
  max_budget = 100000000 # just a very high number, as max budget limits until when the intervals fire.
  receivers  = [gamefabric_notification_receiver.accounting.name]
  interval = {
    start = "10000"
    step  = "10000"
  }
}
