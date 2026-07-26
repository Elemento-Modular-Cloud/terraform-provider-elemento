terraform {
  required_providers {
    elemento = {
      source = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  host     = "azure.dev.meson.elemento.cloud"
  username = "user@elemento.cloud"
  password = "password"
}

resource "elemento_openclaw" "my_openclaw" {
  vm_name           = "my-openclaw"
  region            = "italynorth"
  billing_frequency = "day"
  cpu_cores         = 4
  ram_size_mb       = 8192
  username          = "elemento"
  password          = "elemento123!"
  ssh_key           = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v g1org1o@archsung"
}

output "payment_url" {
  description = "Payment URL for the created resource"
  value       = elemento_openclaw.my_openclaw.payment_url
}
