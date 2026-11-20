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
  name             = "my-privatenetwork"
  cidr             = "192.168.2.0/24"
  auto_assign_cidr = false
  region           = "eu-south-1"
}

resource "elemento_publicgateway" "my_publicgateway" {
  name         = "my-publicgateway"
  region       = "eu-south-1"
  network_uuid = elemento_privatenetwork.my_privatenetwork.uuid

  depends_on = [
    elemento_privatenetwork.my_privatenetwork
  ]
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_objectstorage.my_bucket.payment_url
# }
