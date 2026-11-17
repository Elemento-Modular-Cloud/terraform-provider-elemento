package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKAASResource(t *testing.T) {
	const resourceName = "elemento_kaas.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_kaas" "test" {
  name              = %[1]q
  region            = "fi-hel2"
  billing_frequency = "day"
  node_pools = [
    {
      name              = "my-node-pool"
      auto_scale        = false
      min_nodes         = 3
      max_nodes         = 3
      shared_storage_gb = 0
      specs = {
        cpu_cores    = 2
        memory_gb    = 4
        disk_size_gb = 64
        gpus         = 0
      }
    }
  ]
  nodes_subnet = {
    dhcp_enabled = true
    cidr         = "192.170.2.0/24"
  }
  update_policy            = "ALWAYS_UPDATE"
  kubernetes_version       = "1.35"
  control_plane_ha_enabled = true
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "node_pools.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "kubernetes_version", "1.35"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
