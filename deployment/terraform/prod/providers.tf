terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.16.0"
    }
  }
}

provider "google" {
  alias = "impersonation"
  scopes = [
    "https://www.googleapis.com/auth/cloud-platform"
  ]
}

data "google_service_account_access_token" "default" {
  provider               = google.impersonation
  target_service_account = var.terraform_service_account
  lifetime               = "900s"

  scopes = ["cloud-platform"]
}

provider "google" {
  project = var.project_id
  region  = var.region
}