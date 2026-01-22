output "vpc_network_id" {
  description = "The id of the internal network used for 'app -> redis' communication"
  value       = google_compute_network.vpc.id
}

output "vpc_connector_name" {
  description = "The connector name of the internal network used for 'app -> redis' secure connection"
  value       = google_vpc_access_connector.connector.name
}

output "cloud_sql_vpc_id" {
  description = "Private IP address for PostgreSQL"
  value       = google_compute_network.vpc.id
}