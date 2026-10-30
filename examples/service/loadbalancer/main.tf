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

resource "elemento_loadbalancer" "my_loadbalancer" {
  name      = "my-loadbalancer"
  type      = "application"
  is_public = false
  port      = 80
  region    = "eu-south-1"
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_objectstorage.my_bucket.payment_url
# }
