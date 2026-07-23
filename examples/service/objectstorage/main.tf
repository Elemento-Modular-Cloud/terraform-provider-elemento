terraform {
  required_providers {
    elemento = {
      source  = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  host     = "azure.dev.meson.elemento.cloud"
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
