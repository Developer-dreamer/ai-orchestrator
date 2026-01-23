// configured from: https://cloud.google.com/run/docs/samples/cloudrun-connect-cloud-sql-parent-tag
resource "google_project_service" "sqladmin_api" {
  service            = "sqladmin.googleapis.com"
  disable_on_destroy = false
}

resource "google_sql_database_instance" "default" {
  name             = "${var.service_name}-sql-db"
  region           = var.region
  database_version = var.database_version
  root_password    = var.db_password


  settings {
    activation_policy = "ALWAYS"
    tier = var.tier

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.vpc_id
    }

    disk_size = 11
    disk_type = "PD_SSD"

    backup_configuration {
      enabled                        = true
      binary_log_enabled             = false
      start_time                     = "03:00"
      location                       = "us"
      transaction_log_retention_days = 7
    }

    availability_type = "ZONAL"
  }

  deletion_protection = false
  depends_on = [google_project_service.sqladmin_api]
}

resource "google_sql_database" "default" {
  name     = var.db_name
  instance = google_sql_database_instance.default.name
}