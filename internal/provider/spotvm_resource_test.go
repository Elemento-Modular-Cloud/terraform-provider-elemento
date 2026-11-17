package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccSpotVMResource(t *testing.T) {
	const resourceName = "elemento_spotvm.test"
	const resourceFormat = `
resource "elemento_spotvm" "test" {
  vm_name = %[1]q
  cpu = {
    slots        = %[2]d
    shared_cores = true
    arch         = ["X86_64"]
  }
  mem = {
    capacity_mb = %[3]d
  }
  os = {
    family  = "linux"
    flavour = "ubuntu"
    version = "24.04"
  }
  storage = {
    boot = {
      name    = "root"
      size_gb = 32
    }
    data = []
  }
  auth = {
    username = "elemento"
    password = "elemento123"
    ssh_key  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v"
  }
  tags          = %[4]s
  region        = %[5]q
  cloudinit_b64 = "I2Nsb3VkLWNvbmZpZwpwYWNrYWdlczoKICAtIG5naW54CnJ1bmNtZDoKICAtIFsgc3lzdGVtY3RsLCBlbmFibGUsIG5naW54IF0K"
}
`

	testAccCreateImport(
		t,
		resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(
				resourceFormat,
				name,
				2,
				4096,
				`[]`,
				"eu-south-1",
			)
		},
		func(name string) []resource.TestCheckFunc{
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "vm_name", name),

				// CPU nested attributes
				resource.TestCheckResourceAttr(resourceName, "cpu.slots", "2"),
				resource.TestCheckResourceAttr(resourceName, "cpu.shared_cores", "true"),
				resource.TestCheckResourceAttr(resourceName, "cpu.arch.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "cpu.arch.0", "X86_64"),

				// Memory nested attributes
				resource.TestCheckResourceAttr(resourceName, "mem.capacity_mb", "4096"),

				// OS nested attributes
				resource.TestCheckResourceAttr(resourceName, "os.family", "linux"),
				resource.TestCheckResourceAttr(resourceName, "os.flavour", "ubuntu"),
				resource.TestCheckResourceAttr(resourceName, "os.version", "24.04"),

				// Storage nested attributes
				resource.TestCheckResourceAttr(resourceName, "storage.boot.name", "root"),
				resource.TestCheckResourceAttr(resourceName, "storage.boot.size_gb", "32"),
				resource.TestCheckResourceAttr(resourceName, "storage.data.#", "0"),

				// Auth nested attributes
				resource.TestCheckResourceAttr(resourceName, "auth.username", "elemento"),
				resource.TestCheckResourceAttr(resourceName, "auth.password", "elemento123"),
				resource.TestCheckResourceAttr(resourceName, "auth.ssh_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v"),

				// Top-level attributes
				resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				resource.TestCheckResourceAttr(resourceName, "region", "eu-south-1"),
				resource.TestCheckResourceAttr(resourceName, "cloudinit_b64", "I2Nsb3VkLWNvbmZpZwpwYWNrYWdlczoKICAtIG5naW54CnJ1bmNtZDoKICAtIFsgc3lzdGVtY3RsLCBlbmFibGUsIG5naW54IF0K"),
			}
		},
		[]string{
			"auth.password",
			"auth.username",
			"auth.ssh_key",
			"payment_url",
			"status",
		},
	)
}
