resource "google_project_service" "vpc_api" {
  service            = "vpcaccess.googleapis.com"
  project            = var.project_id
  disable_on_destroy = false
}

resource "google_project_service" "service_networking_api" {
  project            = var.project_id
  service            = "servicenetworking.googleapis.com"
  disable_on_destroy = false
}

resource "google_compute_network" "vpc" {
  name                    = "${var.service_name}-vpc"
  auto_create_subnetworks = false
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network = google_compute_network.vpc.id
  service = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_range.name]

  depends_on = [google_project_service.service_networking_api]
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.service_name}-subnet"
  ip_cidr_range = "10.10.10.0/24"
  region        = var.region
  network       = google_compute_network.vpc.id
}

resource "google_vpc_access_connector" "connector" {
  name          = "${var.service_name}-connector-for-redis"
  region        = var.region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.8.0.0/28"

  min_throughput = 200
  max_throughput = 500

  depends_on = [google_project_service.vpc_api]
}

resource "google_compute_global_address" "private_ip_range" {
  name          = "${var.service_name}-private-ip-range"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
}