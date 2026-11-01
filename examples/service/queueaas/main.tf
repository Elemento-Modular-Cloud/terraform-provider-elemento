terraform {
  required_providers {
    elemento = {
      source = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  target   = "aws-dev-private"
  username = "user@elemento.cloud"
  password = "password"
}

resource "elemento_privatenetwork" "my_privatenetwork" {
  name              = "my-privatenetwork"
  cidr              = "192.168.2.0/24"
  auto_assign_cidr  = false
  region            = "eu-south-1"
}

resource "elemento_queueaas" "my_queueaas" {
  name   = "my-queueaas"
  region = "eu-south-1"
  node_number = 3
  node_size = "M"
  size_gb = 10
  username = "user"
  password = "1234515438u_sdf!"
  kafka_version = "3.9.1"
  is_ha = true
  is_public = false
  network_uuid = elemento_privatenetwork.my_privatenetwork.vpc_id

  depends_on = [
    elemento_privatenetwork.my_privatenetwork
  ]
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_objectstorage.my_bucket.payment_url
# }
