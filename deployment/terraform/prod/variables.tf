variable "project_id" {
  description = "The ID of the GCP project"
  type        = string
}

variable "service_name" {
  description = "The name of the service"
  type        = string
  default     = "ai-orchestrator"
}

variable "region" {
  description = "The region for the GCP resources"
  type        = string
  default     = "us-central1"
}

variable "terraform_service_account" {
  description = "The service account used for Terraform operations"
  type        = string
  default     = "ai-orchestrator@poised-graph-484915-a3.iam.gserviceaccount.com"
}

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

variable "app_version" {
  description = "Current version of deploy"
  type        = string
}