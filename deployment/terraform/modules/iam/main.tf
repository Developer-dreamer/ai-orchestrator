resource "google_project_service" "iam" {
  service            = "iam.googleapis.com"
  disable_on_destroy = false
}

resource "google_service_account" "cloud_run" {
  account_id   = "${var.service_name}-cloud-run"
  display_name = "Service account for Cloud Run"
  depends_on = [google_project_service.iam]
}

resource "google_project_iam_member" "cloud_run" {
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
  project = var.project_id
  role    = "roles/run.admin"
}