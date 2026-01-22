# AI-Orchestrator - Terraform configuration

## Quick overview

This module introduces deployment of the app to the GCP using terraform.  
You can find here next configuration directories:
```
terraform
├── modules
│        ├── cloud_run
│        │       ├── api
│        │       └── worker
│        ├── cloud_sql
│        ├── iam
│        ├── memory_store
│        ├── secrets
│        └── vpc
└── prod
    ├── config
    └── main.tf
```

This configuration sets up two [Cloud Run](../modules/cloud_run) instances (in other words, independent Docker containers that run similarly to the usual ones on a local machine).
The first one is for [API](../modules/cloud_run/api) and the second one for [Worker](../modules/cloud_run/worker) microservices. As you know from [main README](../../../README.md), those services communicate via **Redis Streams**.
So, of course, there is a [Redis instance](../modules/memory_store) and a [VPC network](../modules/vpc) for private communication between those microservices and the Redis instance. There is also running [CloudSQL](../modules/cloud_sql) service with underlying **PostgreSQL-15** instance.
And in the end, there is a configured [IAM](../modules/iam) service to manage resource access, and [Secrets](../modules/secrets) to store sensitive information securely.

## Requirements

First of all, you must have a Google Cloud account. And on this account must be a project with set-up billing. 
You also need an installed **gcloud CLI** and **terraform CLI**, to run commands:
```bash
brew install --cask google-cloud-sdk

brew tap hashicorp/tap
brew install hashicorp/tap/terraform
```

Then, you need to setup a `.terraform.env` file with required variables:
```dotenv
TF_VAR_project_id=<your-project-id>
TF_VAR_region=us-central1
TF_VAR_terraform_api_service_account=ai-orchestrator-api@<your-project-id>.iam.gserviceaccount.com
TF_VAR_terraform_worker_service_account=ai-orchestrator-worker@<your-project-id>.iam.gserviceaccount.com
TF_VAR_repo_name=ai-orchestrator
TF_VAR_app_version=v1.0.0

TF_VAR_db_name=orchestrator_db
TF_VAR_db_user=<your-user>
TF_VAR_db_password=<your-secret-password> # use here letters (lower/uppercase) and numbers only

TF_VAR_gemini_api_key=<your-secret-apikey>
```

After you filled the `.terraform.env` file, run:
```bash
source .terraform.env
```
and then validate the result by running:
```bash
echo $TF_VAR_project_id
```
It should return the project ID you have specified in the file.

After this step, you should create a repository inside **Artifacts registry**. Here is the command you should run:
```bash
gcloud artifacts repositories create $TF_VAR_repo_name \
    --repository-format=docker \
    --location=$TF_VAR_region \
    --description="Docker repository for AI Orchestrator"
```

Now, you're almost ready to push the run the deployment, everything you need, is to build and push and image to the **Artifacts registry** by running those commands:
```bash
docker login
```
```bash
docker buildx build \
             --platform linux/amd64 \
             -f ../../docker/Dockerfile \
             --build-arg SERVICE_PATH=cmd/api \
             -t $TF_VAR_region-docker.pkg.dev/$TF_VAR_project_id/$TF_VAR_repo_name/api:$TF_VAR_app_version \
             --push \
             ../../../
```
```bash
docker buildx build \
             --platform linux/amd64 \
             -f ../../docker/Dockerfile \
             --build-arg SERVICE_PATH=cmd/worker \
             -t $TF_VAR_region-docker.pkg.dev/$TF_VAR_project_id/$TF_VAR_repo_name/worker:$TF_VAR_app_version \
             --push \
             ../../../
```
Now, you're ready to start actually deploying app.

## Deployment

The first step: Login into gcloud CLI:
```bash
gcloud auth application-default login
```
this will redirect you to browser, and you must select the account you're actually logged in on Google Cloud.

Look carefully into your login logs, the Google should specify in which project environment you're currently in. It must be the same as you're **TF_VAR_project_id** value.

Then run the next commands to grant your account required permissions to deploy the app on GCP:
```bash
gcloud projects add-iam-policy-binding $TF_VAR_project_id \
  --member="user:$(gcloud config get-value account)" \
  --role="roles/iam.serviceAccountTokenCreator"
```
```bash
gcloud projects add-iam-policy-binding $TF_VAR_project_id \
  --member="user:$(gcloud config get-value account)" \
  --role="roles/storage.objectUser"
```

Now, the actually **terraform** command:
```bash
terraform init
```
```bash
terraform plan
```
>[!NOTE]
> Look carefully whether terraform wants to apply the thing you want. 
> But since this configuration is ready, you should not worry about it.

And the final command:
```bash
terraform apply
```
It will prompt you whether you are sure you want to apply those changes, so just type yes, and that's it.

