output "memstore_connection_string" {
  description = "The redis connection string used to connect app to the redis"
  value       = "redis://:${google_redis_instance.memstore.auth_string}@${google_redis_instance.memstore.host}:${google_redis_instance.memstore.port}"
  sensitive   = true
}

output "memstore_name" {
  description = "The redis instance name  used to save TLS certificate into secret manager and the connect app to redis"
  value       = google_redis_instance.memstore.name
}