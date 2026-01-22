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

  secret_data = templatefile("${path.module}/../../../prod/config/api.yaml.tftpl", {
    service_name            = var.service_name
    environment             = var.environment
    db_host                 = "/cloudsql/${var.db_connection_name}"
    db_user                 = var.db_user
    db_name                 = var.db_name
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
    min_instance_count = 0
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
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.repo_name}/api:${var.app_version}"

      resources {
        limits = {
          "cpu"    = "2"
          "memory" = "2Gi"
        }
      }

      env {
        name = "POSTGRES_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = var.db_pass_secret_id
            version = "latest"
          }
        }
      }
      env {
        name  = "YAML_CFG_DIR"
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
        name       = "config-vol"
        mount_path = "/etc/secrets"
      }
    }

    volumes {
      name = "config-vol"
      secret {
        secret = var.app_config_secret_id
        items {
          version = "latest"
          path    = "config.yaml"
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

  depends_on = [google_secret_manager_secret_version.app_config]
}

resource "google_cloud_run_v2_service_iam_member" "public_access" {
  member   = "allUsers"
  name     = google_cloud_run_v2_service.backend.name
  role     = "roles/run.invoker"
  location = google_cloud_run_v2_service.backend.location
}