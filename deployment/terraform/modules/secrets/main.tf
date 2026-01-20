resource "google_project_service" "secretmanager_api" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_secret_manager_secret" "dbuser" {
  secret_id = "dbuser"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "dbuser_data" {
  secret      = google_secret_manager_secret.dbuser.id
  secret_data = var.db_user
}

resource "google_secret_manager_secret_iam_member" "secretaccess_compute_dbuser" {
  secret_id = google_secret_manager_secret.dbuser.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account_email}"
}

resource "google_secret_manager_secret" "dbpass" {
  secret_id = "dbpass"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "dbpass_data" {
  secret      = google_secret_manager_secret.dbpass.id
  secret_data = var.db_password
}

resource "google_secret_manager_secret_iam_member" "secretaccess_compute_dbpass" {
  secret_id = google_secret_manager_secret.dbpass.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account_email}"
}

resource "google_secret_manager_secret" "dbname" {
  secret_id = "dbname"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "dbname_data" {
  secret      = google_secret_manager_secret.dbname.id
  secret_data = var.db_name
}

resource "google_secret_manager_secret_iam_member" "secretaccess_compute_dbname" {
  secret_id = google_secret_manager_secret.dbname.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account_email}"
}

data "google_redis_instance" "cache" {
  name   = var.memstore_name
  region = var.region
}

resource "google_secret_manager_secret" "redis_ca" {
  secret_id = "redis_ca"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "redis_ca_version" {
  secret      = google_secret_manager_secret.redis_ca.id
  secret_data = data.google_redis_instance.cache.server_ca_certs[0].cert
}

resource "google_secret_manager_secret_iam_member" "secretaccess_compute_redis" {
  secret_id = google_secret_manager_secret.redis_ca.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account_email}"
}
