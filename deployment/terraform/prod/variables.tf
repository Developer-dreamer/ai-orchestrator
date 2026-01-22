# ===== TERRAFORM CONFIGURATION =====

variable "project_id" {
  description = "The ID of the GCP project"
  type        = string
}

variable "api_service_name" {
  description = "The name of the service"
  type        = string
  default     = "ai-orchestrator-api"
}

variable "worker_service_name" {
  description = "The name of the service"
  type        = string
  default     = "ai-orchestrator-worker"
}

variable "region" {
  description = "The region for the GCP resources"
  type        = string
  default     = "us-central1"
}

variable "terraform_api_service_account" {
  description = "The service account used for Terraform operations"
  type        = string
}

variable "terraform_worker_service_account" {
  description = "The service account used for Terraform operations"
  type        = string
}

variable "repo_name" {
  description = "Artifacts repository where the image is located"
  type        = string
  default     = "ai-orchestrator"
}

variable "app_version" {
  description = "Current version of deploy"
  type        = string
}

# ===== COMMON =====

variable "environment" {
  description = "The environment app is running in"
  type        = string
  default     = "production"
}

# ===== API =====

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


# ===== WORKER =====

variable "number_of_workers" {
  description = "The number of workers used to process AI requests"
  type        = number
  default     = 5
}

variable "gemini_api_key" {
  description = "The key used to make API requests to Gemini models"
  type        = string
  sensitive   = true
}