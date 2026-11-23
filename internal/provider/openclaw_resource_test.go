package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type openclawTestParams struct {
	cpuCores int
	ramSizeMB int
	diskSizeGB int
	username string
	password string
	sshKey string
}

func TestAccOpenClawResource(t *testing.T) {
	paramsByProvider := map[string]n8nTestParams{
		"aws": {
			cpuCores: 4,
			ramSizeMB: 8192,
			diskSizeGB: 64,
			username: "elemento",
			password: "elemento123!",
			sshKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v",
		},
	}
	const resourceType = "openclaw"
	const resourceName = "elemento_openclaw.test"
	const resourceFormat = `
resource "elemento_openclaw" "test" {
  vm_name           = %[1]q
  cpu_cores         = %[2]d
  ram_size_mb       = %[3]d
  disk_size_gb      = %[4]d
  username          = %[5]q
  password          = %[6]q
  ssh_key           = %[7]q
  region            = %[8]q
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
				p.cpuCores,
				p.ramSizeMB,
				p.diskSizeGB,
				p.username,
				p.password,
				p.sshKey,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "cpu_cores", strconv.Itoa(p.cpuCores)),
				resource.TestCheckResourceAttr(resourceName, "ram_size_mb", strconv.Itoa(p.ramSizeMB)),
				resource.TestCheckResourceAttr(resourceName, "disk_size_gb", strconv.Itoa(p.diskSizeGB)),
				resource.TestCheckResourceAttr(resourceName, "username", p.username),
				resource.TestCheckResourceAttr(resourceName, "password", p.password),
				resource.TestCheckResourceAttr(resourceName, "ssh_key", p.sshKey),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
			}
		},
		[]string{"username", "password", "ssh_key", "payment_url", "status", "last_updated", "state", "service_url"},
	)
}
