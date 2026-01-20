output "db_connection_name" {
  description = "Connection name used to find the PostgreSQL instance in cloud network and connect to it. Cloud Run does not use such IPs as localhost"
  value       = google_sql_database_instance.default.connection_name
  sensitive   = false
}

output "db_private_ip" {
  description = "Connection IP address used to connect the PostgreSQL instance in cloud network and connect to it. Cloud Run does not use such IPs as localhost"
  value       = google_sql_database_instance.default.name
}