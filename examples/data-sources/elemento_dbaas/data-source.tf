data "elemento_dbaas" "app" {
  uuid = var.existing_db_uuid
}

output "db_host" {
  value = data.elemento_dbaas.app.database_host
}
