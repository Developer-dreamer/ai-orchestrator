resource "google_project_service" "redis_api" {
  service            = "redis.googleapis.com"
  disable_on_destroy = false
}

resource "google_redis_instance" "memstore" {
  name                    = "redis"
  region                  = var.region
  auth_enabled            = true
  memory_size_gb          = 1
  authorized_network      = var.authorized_network_id
  transit_encryption_mode = "SERVER_AUTHENTICATION"

  persistence_config {
    persistence_mode    = "RDB"
    rdb_snapshot_period = "ONE_HOUR"
  }

  depends_on = [google_project_service.redis_api]
}


