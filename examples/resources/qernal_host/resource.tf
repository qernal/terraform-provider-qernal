# Manage example qernal_host.
resource "qernal_host" "host_example" {
  project_id  = "4cab71f8-19da-4bc9-8ee4-6bcc46c1adf0"
  name        = "host_name"
  certificate = "projects:secrets/MY_CERT"
  disabled    = true
}
