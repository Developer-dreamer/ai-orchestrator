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
  description = "The secret ID of the PostgreSQL used to extract database password for connection"
  value       = google_secret_manager_secret.dbpass.id
}

output "api_config_id" {
  description = "The ID used to get yaml configuration file"
  value       = google_secret_manager_secret.api_config.id
}

output "worker_config_id" {
  description = "The ID used to get yaml configuration file"
  value       = google_secret_manager_secret.worker_config.id
}

output "gemini_api_key_secret_id" {
  description = "The key used to make API requests to Gemini models"
  value       = google_secret_manager_secret.gemini_api_key.id
}

output "otel_resource" {
  description = "The Secret Manager ID containing OpenTelemetry resource attributes (e.g., service.name=my-app)"
  value       = google_secret_manager_secret.otel_resource.id
}

output "otel_endpoint" {
  description = "The Secret Manager ID containing the OpenTelemetry OTLP endpoint URL (e.g., https://otlp-gateway...)"
  value       = google_secret_manager_secret.otel_endpoint.id
}

output "otel_headers" {
  description = "The Secret Manager ID containing OpenTelemetry exporter headers, specifically the Authorization token"
  value       = google_secret_manager_secret.otel_headers.id
}