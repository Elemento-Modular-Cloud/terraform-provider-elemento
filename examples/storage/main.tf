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
  region            = "fi-hel2"
  billing_frequency = "month"
}
