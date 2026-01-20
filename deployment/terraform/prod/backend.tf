terraform {
  backend "gcs" {
    bucket = "orchestrator-bucket"
  }
}