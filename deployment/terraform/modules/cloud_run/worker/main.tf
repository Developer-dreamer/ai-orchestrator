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

resource "google_secret_manager_secret_version" "app_config" {
  secret = var.app_config_secret_id

  secret_data = templatefile("${path.module}/../../../prod/config/worker.yaml.tftpl", {
    service_name         = var.service_name
    environment          = var.environment
    redis_host           = var.redis_host
    number_of_workers    = var.number_of_workers
    redis_pub_stream_id  = "results"
    redis_sub_stream_id  = "tasks"
    redis_consumer_group = "ai_tasks_group"
    worker_id            = "worker"
    otel_collector_uri   = "otel-collector:4318"
  })
}

resource "google_cloud_run_v2_service" "backend" {
  location = var.region
  name     = "${var.service_name}-worker"
  ingress  = local.traffic_type

  labels = {
    "app" = "${var.service_name}-worker"
  }

  scaling {
    min_instance_count = 1
  }

  deletion_protection = false

  template {
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
        name  = "GEMINI_API_KEY"
        value = data.google_secret_manager_secret_version.gemini_api_key.secret_data
      }
      env {
        name  = "YAML_CFG_PATH"
        value = "/etc/secrets/config.yaml"
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
