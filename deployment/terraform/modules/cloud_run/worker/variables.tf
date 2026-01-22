# ===== TERRAFORM ======

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

variable "app_version" {
  description = "Current version of deploy"
  type        = string
}

variable "repo_name" {
  description = "Artifacts repository where the image is located"
  type        = string
}


# ===== APP =====

variable "environment" {
  description = "The environment the app is being run inside"
  type        = string
}

variable "app_config_secret_id" {
  description = "The secret id of app configuration"
  type        = string
}

variable "number_of_workers" {
  description = "The amount of workers to process AI requests"
  type        = string
}

variable "gemini_api_key_secret_id" {
  description = "The key used to make API requests to Gemini models"
  type        = string
}


# ===== REDIS =====

variable "redis_host" {
  description = "The host of Redis instance"
  type        = string
}

variable "redis_secret_id" {
  description = "The secret ID of the redis used to extract TLS certificate for secure connection"
  type        = string
}


# ===== VPC =====

variable "vpc_connector_name" {
  description = "The connector name of the internal network used for 'app -> redis' secure connection"
  type        = string
}
