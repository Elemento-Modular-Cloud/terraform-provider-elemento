terraform {
  required_providers {
    elemento = {
      source = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  target   = "scaleway-dev-private"
  username = "user@elemento.cloud"
  password = "password"
}

resource "elemento_openclaw" "my_openclaw" {
  vm_name           = "my-openclaw"
  region            = "it-mil"
  billing_frequency = "day"
  cpu_cores         = 4
  ram_size_mb       = 8192
  disk_size_gb      = 64
  username          = "elemento"
  password          = "elemento123!"
  ssh_key           = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v g1org1o@archsung"
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_openclaw.my_openclaw.payment_url
# }
