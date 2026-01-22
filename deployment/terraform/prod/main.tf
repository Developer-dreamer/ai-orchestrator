module "iam_api" {
  source       = "../modules/iam"
  project_id   = var.project_id
  service_name = var.api_service_name

  region                = var.region
  service_account_email = var.terraform_api_service_account
}

module "iam_worker" {
  source       = "../modules/iam"
  project_id   = var.project_id
  service_name = var.worker_service_name

  region                = var.region
  service_account_email = var.terraform_worker_service_account
}

module "vpc" {
  source     = "../modules/vpc"
  project_id = var.project_id
  region     = var.region
}

module "cloud_sql" {
  source       = "../modules/cloud_sql"
  service_name = var.api_service_name
  db_password  = var.db_password
  db_name      = var.db_name
  region       = var.region
  vpc_id       = module.vpc.cloud_sql_vpc_id

  depends_on = [module.vpc]
}

module "memory_store" {
  source                = "../modules/memory_store"
  authorized_network_id = module.vpc.vpc_network_id
  region                = var.region

  depends_on = [module.vpc]
}

module "secrets" {
  source                       = "../modules/secrets"
  api_service_account_email    = module.iam_api.cloud_run_service_account_email
  worker_service_account_email = module.iam_worker.cloud_run_service_account_email
  db_name                      = var.db_name
  db_user                      = var.db_user
  db_password                  = var.db_password
  region                       = var.region
  redis_ca_cert                = module.memory_store.server_ca_certs[0].cert
  gemini_api_key               = var.gemini_api_key

  depends_on = [module.iam_api, module.iam_worker, module.memory_store]
}


module "api" {
  source                = "../modules/cloud_run/api"
  service_name          = var.api_service_name
  project_id            = var.project_id
  app_version           = var.app_version
  service_account_email = module.iam_api.cloud_run_service_account_email
  region                = var.region
  repo_name             = var.repo_name
  environment           = var.environment
  app_config_secret_id  = module.secrets.api_config_id

  db_connection_name = module.cloud_sql.db_connection_name
  db_name            = var.db_name
  db_pass_secret_id  = module.secrets.db_pass_secret_id
  db_user            = var.db_user

  redis_host      = module.memory_store.memstore_connection_string
  redis_secret_id = module.secrets.redis_secret_id

  vpc_connector_name = module.vpc.vpc_connector_name

  depends_on = [
    module.cloud_sql,
    module.memory_store,
    module.secrets
  ]
}

module "worker" {
  source                = "../modules/cloud_run/worker"
  service_name          = var.worker_service_name
  project_id            = var.project_id
  app_version           = var.app_version
  service_account_email = module.iam_worker.cloud_run_service_account_email
  region                = var.region
  repo_name             = var.repo_name
  environment           = var.environment
  app_config_secret_id  = module.secrets.worker_config_id
  number_of_workers     = var.number_of_workers
  redis_host            = module.memory_store.memstore_connection_string
  redis_secret_id       = module.secrets.redis_secret_id

  vpc_connector_name = module.vpc.vpc_connector_name

  gemini_api_key_secret_id = module.secrets.gemini_api_key_secret_id

  depends_on = [
    module.cloud_sql,
    module.memory_store,
    module.secrets
  ]
}
