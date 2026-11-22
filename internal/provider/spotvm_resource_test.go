package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type spotVMTestParams struct {
	slots         int
	sharedCores   bool
	arch          string
	capacityMB    int
	osFamily      string
	osFlavour     string
	osVersion     string
	sizeGB        int
	username      string
	password      string
	sshKey        string
	tags          string
	cloudinit_b64 string
}

func TestAccSpotVMResource(t *testing.T) {
	paramsByProvider := map[string]spotVMTestParams{
		"aws": {
			slots:         2,
			sharedCores:   true,
			arch:          "X86_64",
			capacityMB:    4096,
			osFamily:      "linux",
			osFlavour:     "ubuntu",
			osVersion:     "24.04",
			sizeGB:        32,
			username:      "elemento",
			password:      "elemento123",
			sshKey:        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v",
			tags:          "[]",
			cloudinit_b64: "I2Nsb3VkLWNvbmZpZwpwYWNrYWdlczoKICAtIG5naW54CnJ1bmNtZDoKICAtIFsgc3lzdGVtY3RsLCBlbmFibGUsIG5naW54IF0K",
		},
	}

	const resourceType = "spotvm"
	const resourceName = "elemento_spotvm.test"
	const resourceFormat = `
resource "elemento_spotvm" "test" {
  vm_name = %[1]q
  cpu = {
    slots        = %[2]d
    shared_cores = %[3]t
    arch         = [%[4]q]
  }
  mem = {
    capacity_mb = %[5]d
  }
  os = {
    family  = %[6]q
    flavour = %[7]q
    version = %[8]q
  }
  storage = {
    boot = {
      name    = "root"
      size_gb = %[9]d
    }
    data = []
  }
  auth = {
    username = %[10]q
    password = %[11]q
    ssh_key  = %[12]q
  }
  tags          = %[13]s
  cloudinit_b64 = %[14]q
  region        = %[15]q
}
`

	testAccCreateImport(
		t,
		resourceName,
		resourceType,
		func(name string, target ProviderConfig) string {
			p := paramsByProvider[target.name]

			return target.ProviderBlock() + fmt.Sprintf(
				resourceFormat,
				name,
				p.slots,
				p.sharedCores,
				p.arch,
				p.capacityMB,
				p.osFamily,
				p.osFlavour,
				p.osVersion,
				p.sizeGB,
				p.username,
				p.password,
				p.sshKey,
				p.tags,
				p.cloudinit_b64,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "vm_name", name),

				// CPU nested attributes
				resource.TestCheckResourceAttr(resourceName, "cpu.slots", strconv.Itoa(p.slots)),
				resource.TestCheckResourceAttr(resourceName, "cpu.shared_cores", strconv.FormatBool(p.sharedCores)),
				resource.TestCheckResourceAttr(resourceName, "cpu.arch.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "cpu.arch.0", p.arch),

				// Memory nested attributes
				resource.TestCheckResourceAttr(resourceName, "mem.capacity_mb", strconv.Itoa(p.capacityMB)),

				// OS nested attributes
				resource.TestCheckResourceAttr(resourceName, "os.family", p.osFamily),
				resource.TestCheckResourceAttr(resourceName, "os.flavour", p.osFlavour),
				resource.TestCheckResourceAttr(resourceName, "os.version", p.osVersion),

				// Storage nested attributes
				resource.TestCheckResourceAttr(resourceName, "storage.boot.name", "root"),
				resource.TestCheckResourceAttr(resourceName, "storage.boot.size_gb", strconv.Itoa(p.sizeGB)),
				resource.TestCheckResourceAttr(resourceName, "storage.data.#", "0"),

				// Auth nested attributes
				resource.TestCheckResourceAttr(resourceName, "auth.username", p.username),
				resource.TestCheckResourceAttr(resourceName, "auth.password", p.password),
				resource.TestCheckResourceAttr(resourceName, "auth.ssh_key", p.sshKey),

				// Top-level attributes
				resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),
				resource.TestCheckResourceAttr(resourceName, "cloudinit_b64", p.cloudinit_b64),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
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
