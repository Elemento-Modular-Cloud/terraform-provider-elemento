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

resource "elemento_filestorage" "my_filestorage" {
  name      = "my-filestorage"
  size_gb   = 10
  is_public = false
  region    = "eu-south-1"
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_objectstorage.my_bucket.payment_url
# }
