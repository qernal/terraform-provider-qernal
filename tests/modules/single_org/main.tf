resource "qernal_organisation" "org" {
  name = var.org_name
}

output "organisation_name" {
  value =  qernal_organisation.org.name
}
