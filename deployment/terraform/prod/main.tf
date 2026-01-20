module "iam" {
  source       = "../modules/iam"
  project_id   = var.project_id
  service_name = var.service_name
}

module "vpc" {
  source       = "../modules/vpc"
  project_id   = var.project_id
  region       = var.region
  service_name = var.service_name
}

module "cloud_sql" {
  source       = "../modules/cloud_sql"
  service_name = var.service_name
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
  service_name          = var.service_name

  depends_on = [module.vpc]
}

module "secrets" {
  source                = "../modules/secrets"
  service_account_email = module.iam.cloud_run_service_account_email
  db_name               = var.db_name
  db_user               = var.db_user
  db_password           = var.db_password
  region                = var.region
  memstore_name         = module.memory_store.memstore_name

  depends_on = [module.iam, module.memory_store]
}


module "api" {
  source                     = "../modules/cloud_run/api"
  service_name               = var.service_name
  service_account_email      = module.iam.cloud_run_service_account_email
  region                     = var.region
  db_connection_name         = module.cloud_sql.db_connection_name
  memstore_connection_string = module.memory_store.memstore_connection_string
  vpc_connector_name         = module.vpc.vpc_connector_name
  db_name_secret_id          = module.secrets.db_name_secret_id
  db_pass_secret_id          = module.secrets.db_pass_secret_id
  db_user_secret_id          = module.secrets.db_user_secret_id
  project_id                 = var.project_id
  redis_secret_id            = module.secrets.redis_secret_id
  api_key_secret_id          = module.secrets.api_key_secret_id
  app_version                = var.app_version
  db_private_ip              = module.cloud_sql.db_private_ip

  depends_on = [
    module.cloud_sql,
    module.memory_store,
    module.secrets
  ]
}

module "worker" {
  source                     = "../modules/cloud_run/worker"
  service_name               = var.service_name
  service_account_email      = module.iam.cloud_run_service_account_email
  region                     = var.region
  db_connection_name         = module.cloud_sql.db_connection_name
  memstore_connection_string = module.memory_store.memstore_connection_string
  vpc_connector_name         = module.vpc.vpc_connector_name
  db_name_secret_id          = module.secrets.db_name_secret_id
  db_pass_secret_id          = module.secrets.db_pass_secret_id
  db_user_secret_id          = module.secrets.db_user_secret_id
  project_id                 = var.project_id
  redis_secret_id            = module.secrets.redis_secret_id
  api_key_secret_id          = module.secrets.api_key_secret_id
  app_version                = var.app_version
  db_private_ip              = module.cloud_sql.db_private_ip

  depends_on = [
    module.memory_store,
    module.secrets
  ]
}
