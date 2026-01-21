data "google_secret_manager_secret_version" "gemini_api_key" {
  secret  = var.gemini_api_key_secret_id
  version = "latest"
}

data "google_secret_manager_secret_version" "app_config" {
  secret  = var.app_config_secret_id
  version = "latest"
}
