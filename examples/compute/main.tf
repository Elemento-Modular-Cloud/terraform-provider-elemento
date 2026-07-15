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

resource "elemento_compute" "n8n_cloudinit_test" {
  vm_name           = "n8n-cloudinit-test"
  region            = "fi-hel2"
  billing_frequency = "month"

  requirements = {
    cpu = {
      smt_enabled       = false
      min_frequency     = 1
      arch              = "X86_64"
      flags             = []
      slots             = 2
      allow_smt         = false
      full_physical     = true
      max_overprovision = 1
    }

    memory = {
      capacity   = 4096
      require_ecc = false
    }

    pci = {
      devices = {}
    }

    misc = {
      os_family  = "linux"
      os_flavour = "ubuntu"
    }
  }

  cloudinit = {
    template  = "n8n"
    variables = {}
  }

  additional_info = {
    service = "n8n"
  }

  volumes = [
    {
      name      = "test"
      size      = 32
      private   = true
      shareable = false
      bootable  = false
      readonly  = false
    }
  ]
}

output "payment_url" {
  description = "Payment URL for the created resource (if applicable)"
  value       = elemento_compute.n8n_cloudinit_test.payment_url
}

# output "billing_uuid" {
#   description = "Billing UUID for the created resource"
#   value       = elemento_compute.n8n_cloudinit_test.billing_uuid
# }
