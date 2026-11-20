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

resource "elemento_aiserverless" "my_aiserverless" {
  name             = "my-aiserverless"
  region           = "eu-south-1"
  model_name       = "claude-sonnet-5"
  enable_streaming = true
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_objectstorage.my_bucket.payment_url
# }
