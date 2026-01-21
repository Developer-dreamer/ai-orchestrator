data "google_secret_manager_secret_version" "app_config" {
  secret  = var.app_config_secret_id
  version = "latest"
}
