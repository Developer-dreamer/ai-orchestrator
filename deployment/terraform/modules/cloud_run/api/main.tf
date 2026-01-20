locals {
  traffic_type = "INGRESS_TRAFFIC_ALL"
  postgres_uri = "postgres://${data.google_secret_manager_secret_version.db_user.secret_data}:${data.google_secret_manager_secret_version.db_pass.secret_data}@${data.google_sql_database_instance.default.ip_address[0].ip_address}:5432/${data.google_secret_manager_secret_version.db_name.secret_data}?sslmode=require"
}

resource "google_project_service" "cloud_run_api" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "resource_manager_api" {
  service = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}

data "google_sql_database_instance" "default" {
  name = var.db_private_ip
}

resource "google_cloud_run_v2_service" "backend" {
  location = var.region
  name     = "${var.service_name}-backend"
  ingress  = local.traffic_type

  labels = {
    "app" = var.service_name
  }

  scaling {
    min_instance_count = 1
  }

  deletion_protection = false

  template {
    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [var.db_connection_name]
      }
    }
    volumes {
      name = "redis-ca"
      secret {
        secret = var.redis_secret_id
        items {
          version = "latest"
          path    = "server-ca.pem"
        }
      }
    }


    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.service_name}/app:${var.app_version}"

      resources {
        limits = {
          "cpu"    = "2"
          "memory" = "2Gi"
        }
      }

      env {
        name  = "POSTGRES_URI"
        value = local.postgres_uri
      }
      env {
        name  = "REDIS_URI"
        value = var.memstore_connection_string
      }
      env {
        name  = "APP_ENVIRONMENT"
        value = "Development"
      }
      env {
        name  = "CACHE_TTL_MINUTES"
        value = "5"
      }
      env {
        name  = "CLOUD_PROJECT_ID"
        value = var.project_id
      }


      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
      volume_mounts {
        name       = "redis-ca"
        mount_path = "/certs"
      }
    }

    max_instance_request_concurrency = 20

    vpc_access {
      connector = "projects/${var.project_id}/locations/${var.region}/connectors/${var.vpc_connector_name}"
    }

    service_account = var.service_account_email
  }

  client = "terraform"

  depends_on = [google_project_service.cloud_run_api]
}

resource "google_project_iam_member" "terraform_run_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_cloud_run_v2_service_iam_member" "public_access" {
  member   = "allUsers"
  name     = google_cloud_run_v2_service.backend.name
  role     = "roles/run.invoker"
  location = google_cloud_run_v2_service.backend.location
}

resource "google_project_iam_member" "cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_project_iam_member" "firestore_access" {
  project = var.project_id
  role    = "roles/datastore.user"
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