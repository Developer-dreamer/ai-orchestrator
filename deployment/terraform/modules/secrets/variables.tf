variable "db_name" {
  description = "The database name for CloudSQL PostgreSQL instance"
  type        = string
}

variable "db_user" {
  description = "The user for CloudSQL PostgreSQL instance"
  type        = string
}

variable "db_password" {
  description = "The password for CloudSQL PostgreSQL instance"
  type        = string
  sensitive   = true
}

variable "api_service_account_email" {
  description = "The email of the service account"
  type        = string
}

variable "worker_service_account_email" {
  description = "The email of the service account"
  type        = string
}

variable "region" {
  description = "The region for the GCP resources"
  type        = string
}

variable "memstore_name" {
  description = "The name of the Redis"
  type        = string
}

variable "gemini_api_key" {
  description = "The key used to make API requests to Gemini models"
  type        = string
}

variable "redis_ca_cert" {
  description = "Content of the Redis CA Certificate"
  type        = string
}

