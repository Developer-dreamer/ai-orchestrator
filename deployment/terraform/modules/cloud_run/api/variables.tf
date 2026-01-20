variable "service_name" {
  description = "The name of the service"
  type        = string
}

variable "project_id" {
  description = "The ID of the GCP project"
  type        = string
}

variable "service_account_email" {
  description = "The email of the service account"
  type        = string
}

variable "region" {
  description = "The region for the GCP resources"
  type        = string
}

variable "db_connection_name" {
  description = "Connection name used to find the PostgreSQL instance in cloud network and connect to it. Cloud Run does not use such IPs as localhost"
  type        = string
}

variable "memstore_connection_string" {
  description = "The string to connect to redis"
  type        = string
  sensitive   = true
}

variable "vpc_connector_name" {
  description = "The connector name of the internal network used for 'app -> redis' secure connection"
  type        = string
}

variable "redis_secret_id" {
  description = "The secret ID of the redis used to extract TLS certificate for secure connection"
  type        = string
}

variable "db_user_secret_id" {
  description = "The secret ID of the PostgreSQL used to extract database user for connection"
  type        = string
}

variable "db_name_secret_id" {
  description = "The secret ID of the PostgreSQL used to extract database name for connection"
  type        = string
}

variable "db_pass_secret_id" {
  description = "The secret ID of the PostgreSQL used to extract database name for connection"
  type        = string
}

variable "app_version" {
  description = "Current version of deploy"
  type        = string
}

variable "db_private_ip" {
  description = "IP of the postgres in private network"
  type        = string
}
