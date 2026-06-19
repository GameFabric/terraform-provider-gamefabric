resource "gamefabric_notification_receiver" "accounting" {
  name     = "accounting"
  email_to = ["accounting@example.com"]
}
