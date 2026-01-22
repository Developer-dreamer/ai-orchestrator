resource "google_project_service" "iam" {
  service            = "iam.googleapis.com"
  disable_on_destroy = false
}

resource "google_service_account" "cloud_run" {
  account_id   = var.service_name
  display_name = "Service account for Cloud Run"
  depends_on = [google_project_service.iam]
}

# ===== CHECK IF NECESSARY =====
resource "google_project_iam_member" "terraform_run_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_project_iam_member" "cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_project_iam_member" "service_networking_admin" {
  project = var.project_id
  role    = "roles/servicenetworking.networksAdmin"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_project_iam_member" "resourcemanager_admin" {
  member  = "serviceAccount:${var.service_account_email}"
  project = var.project_id
  role    = "roles/resourcemanager.projectIamAdmin"
}

resource "google_project_iam_member" "serviceusage_admin" {
  member  = "serviceAccount:${var.service_account_email}"
  project = var.project_id
  role    = "roles/serviceusage.serviceUsageAdmin"
}

resource "google_project_iam_member" "terraform_compute_viewer" {
  project = var.project_id
  role    = "roles/compute.admin"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_project_iam_member" "redis_viewer" {
  project = var.project_id
  role    = "roles/redis.viewer"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_project_iam_member" "secrets_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${var.service_account_email}"
}