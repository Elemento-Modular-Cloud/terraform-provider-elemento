package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComputeResource(t *testing.T) {
	const resourceName = "elemento_compute.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_compute" "test" {
  vm_name           = %[1]q
  region            = "fi-hel2"
  billing_frequency = "month"
  requirements = {
    cpu = {
      smt_enabled       = false
      min_frequency     = 1
      arch              = ["X86_64"]
      flags             = []
      slots             = 2
      allow_smt         = false
      full_physical     = true
      max_overprovision = 1
    }
    memory = {
      capacity    = 4096
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
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "vm_name", name),
				resource.TestCheckResourceAttr(resourceName, "requirements.cpu.slots", "2"),
				resource.TestCheckResourceAttr(resourceName, "requirements.memory.capacity", "4096"),
				resource.TestCheckResourceAttr(resourceName, "volumes.#", "1"),
			}
		},
		[]string{"requirements", "cloudinit", "additional_info", "volumes", "authentication", "payment_url", "status"},
	)
}
