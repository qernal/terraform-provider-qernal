data "qernal_organisation" "org" {
  name = var.org_name
}

output "organisation_id" {
  value = data.qernal_organisation.org.id
}

output "organisation_name" {
  value = data.qernal_organisation.org.name
}
