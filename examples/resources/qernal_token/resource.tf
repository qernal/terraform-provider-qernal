# Manage example qernal_token.
resource "qernal_token" "token_example" {
  name            = "ENVIRONMENT_SECRET_NAME_1"
  expiry_duration = 90
}
