terraform {
  required_providers {
    elemento = {
      source = "elemento-modular-cloud/elemento"
    }
  }
}

provider "elemento" {
  target   = "azure-dev-private"
  username = "user@elemento.cloud"
  password = "password"
}

resource "elemento_objectstorage" "my_bucket" {
  name              = "my-bucket"
  size_tb           = 1
  region            = "italynorth"
  billing_frequency = "day"
}

output "payment_url" {
  description = "Payment URL for the created resource"
  value       = elemento_objectstorage.my_bucket.payment_url
}
