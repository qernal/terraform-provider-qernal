data "qernal_project" "project" {
  name = var.project_name
}

output "project_name" {
  value = data.qernal_project.project.name
}

output "project_id" {
  value = data.qernal_project.project.org_id
}
