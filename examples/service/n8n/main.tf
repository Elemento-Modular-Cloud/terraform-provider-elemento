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

resource "elemento_n8n" "my_n8n" {
  vm_name           = "my-n8n"
  region            = "it-mil"
  billing_frequency = "day"
  cpu_cores         = 2
  ram_size_mb       = 4096
  disk_size_gb      = 32
  username          = "elemento"
  password          = "elemento123!"
  ssh_key           = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v g1org1o@archsung"
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_n8n.my_n8n.payment_url
# }
