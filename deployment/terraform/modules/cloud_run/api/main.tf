locals {
  traffic_type = "INGRESS_TRAFFIC_ALL"
}

resource "google_project_service" "cloud_run_api" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "resource_manager_api" {
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}

# data "google_sql_database_instance" "default" {
#   name = var.db_private_ip
# }

resource "google_secret_manager_secret_version" "app_config" {
  secret = var.app_config_secret_id

  secret_data = templatefile("${path.module}/../../../prod/api.yaml.tftpl", {
    service_name            = var.service_name
    environment             = var.environment
    db_host                 = var.db_connection_name
    db_user                 = data.google_secret_manager_secret_version.db_user.secret_data
    db_name                 = data.google_secret_manager_secret_version.db_name.secret_data
    redis_host              = var.redis_host
    api_redis_pub_stream_id = "tasks"
    api_redis_sub_stream_id = "results"
    redis_consumer_group    = "ai_tasks_group"
    worker_id               = "worker"
    otel_collector_uri      = "otel-collector:4318"
  })
}

resource "google_cloud_run_v2_service" "backend" {
  location = var.region
  name     = "${var.service_name}-api"
  ingress  = local.traffic_type

  labels = {
    "app" = "${var.service_name}-api"
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
        name  = "POSTGRES_PASSWORD"
        value = data.google_secret_manager_secret_version.db_pass.secret_data
      }
      env {
        name  = "YAML_CFG_PATH"
        value = "/etc/secrets/config.yaml"
      }

      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
      volume_mounts {
        name       = "redis-ca"
        mount_path = "/certs"
      }
      volume_mounts {
        mount_path = "config-vol"
        name       = "/ect/secrets"
      }
    }

    volumes {
      name = "config-vol"
      secret {
        secret = var.app_config_secret_id
        items {
          key  = "latest"
          path = "config.yaml"
        }
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