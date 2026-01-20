variable "service_name" {
  description = "The name of the service"
  type        = string
}

variable "region" {
  description = "The region where database is deployed"
  type        = string
}

variable "database_version" {
  default     = "POSTGRES_15"
  description = "The version of database used by Cloud SQL"
  type        = string
}

variable "tier" {
  default     = "db-custom-1-3840"
  description = "The instance size"
  type        = string
}

variable "db_password" {
  description = "PostgreSQL password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "PostgreSQL database name"
  type        = string
}

variable "vpc_id" {
  description = "PostgreSQL private IP"
  type        = string
}