variable "service_name" {
  description = "The name of the service"
  type        = string
}

variable "region" {
  description = "The region for the GCP resources"
  type        = string
}

variable "authorized_network_id" {
  type        = string
  description = "The id of the private network used for 'app -> redis' fast and secure communication"
}