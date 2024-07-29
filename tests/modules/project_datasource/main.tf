data "qernal_project" "project" {
  project_id = var.project_id 
}

output "project_name" {
  value = data.qernal_project.project.name 
}
