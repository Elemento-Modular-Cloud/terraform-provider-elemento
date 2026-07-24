terraform {
  required_providers {
    elemento = {
      source  = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  host     = "upcloud.dev.meson.elemento.cloud"
  username = "user@elemento.cloud"
  password = "password"
}

resource "elemento_dbaas" "my_db" {
  name              = "my-db"
  engine            = "postgres"
  disk_size_gb      = 50
  backup_time       = "12:00:00"
  password          = "mysecretpassword"
  nodes_number      = 1
  region            = "fi-hel2"
  billing_frequency = "day"
}

output "payment_url" {
  description = "Payment URL for the created resource"
  value       = elemento_dbaas.my_db.payment_url
}
