output "redis_secret_id" {
  description = "The secret ID of the redis used to extract TLS certificate for secure connection"
  value       = google_secret_manager_secret.redis_ca.id
}

output "db_user_secret_id" {
  description = "The secret ID of the PostgreSQL used to extract database user for connection"
  value       = google_secret_manager_secret.dbuser.id
}

output "db_name_secret_id" {
  description = "The secret ID of the PostgreSQL used to extract database name for connection"
  value       = google_secret_manager_secret.dbname.id
}

output "db_pass_secret_id" {
  description = "The secret ID of the PostgreSQL used to extract database name for connection"
  value       = google_secret_manager_secret.dbpass.id
}
