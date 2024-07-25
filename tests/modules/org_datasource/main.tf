data "qernal_organisation" "org" {
  organisation_id = var.org_id
}

output "organisation_id" {
  value = data.qernal_organisation.org.organisation_id
}
