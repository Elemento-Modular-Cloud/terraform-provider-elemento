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

resource "elemento_storage" "my_boot_volume" {
  name              = "my-boot-volume"
  private           = false
  readonly          = false
  shareable         = false
  bootable          = true
  clonable          = true
  alg               = "cp"
  format            = "qcow2"
  bus               = "virtio"
  size              = 32
  priority          = 100
  tags              = ["test:test"]
  region            = "italynorth"
  billing_frequency = "month"
}

output "payment_url" {
  description = "Payment URL for the created resource"
  value       = elemento_storage.my_boot_volume.payment_url
}
