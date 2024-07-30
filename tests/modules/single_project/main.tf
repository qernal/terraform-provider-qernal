resource "qernal_project" "project" {
  name   = var.project_name
  org_id = var.org_id
}

output "project_name" {
  value = qernal_project.project.name
}

output "project_id" {
  value = qernal_project.project.id
}
