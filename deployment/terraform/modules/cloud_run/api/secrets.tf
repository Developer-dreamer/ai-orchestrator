data "google_secret_manager_secret_version" "db_user" {
  secret  = var.db_user_secret_id
  version = "latest"
}

data "google_secret_manager_secret_version" "db_pass" {
  secret  = var.db_pass_secret_id
  version = "latest"
}

data "google_secret_manager_secret_version" "db_name" {
  secret  = var.db_name_secret_id
  version = "latest"
}

data "google_secret_manager_secret_version" "app_config" {
  secret = var.app_config_secret_id
  version = "latest"
}
