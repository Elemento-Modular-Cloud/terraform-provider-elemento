package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOpenClawResource(t *testing.T) {
	const resourceName = "elemento_openclaw.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_openclaw" "test" {
  vm_name           = %[1]q
  region            = "it-mil"
  billing_frequency = "day"
  cpu_cores         = 4
  ram_size_mb       = 8192
  disk_size_gb      = 64
  username          = "elemento"
  password          = "elemento123!"
  ssh_key           = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "vm_name", name),
				resource.TestCheckResourceAttr(resourceName, "cpu_cores", "4"),
				resource.TestCheckResourceAttr(resourceName, "ram_size_mb", "8192"),
				resource.TestCheckResourceAttr(resourceName, "disk_size_gb", "64"),
			}
		},
		[]string{"password", "ssh_key", "payment_url", "status"},
	)
}
