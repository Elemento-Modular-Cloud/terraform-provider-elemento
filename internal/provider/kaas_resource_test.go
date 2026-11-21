package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type kaasTestParams struct {
	nodePoolAutoScale       bool
	nodePoolMinNodes        int
	nodePoolMaxNodes        int
	nodePoolSharedStorageGB int
	nodePoolCPUCores        int
	nodePoolMemoryGB        int
	nodePoolDiskSizeGB      int
	nodePoolGPUs            int
	dhcpEnabled             bool
	cidr                    string
	updatePolicy            string
	kubernetesVersion       string
	controlPlaneHAEnabled   bool
}

func TestAccKAASResource(t *testing.T) {
	paramsByProvider := map[string]kaasTestParams{
		"aws": {
			nodePoolAutoScale:       false,
			nodePoolMinNodes:        3,
			nodePoolMaxNodes:        3,
			nodePoolSharedStorageGB: 0,
			nodePoolCPUCores:        2,
			nodePoolMemoryGB:        4,
			nodePoolDiskSizeGB:      64,
			nodePoolGPUs:            0,
			dhcpEnabled:             true,
			cidr:                    "192.168.0.0/24",
			updatePolicy:            "ALWAYS_UPDATE",
			kubernetesVersion:       "1.35",
			controlPlaneHAEnabled:   true,
		},
	}

	const resourceType = "kaas"
	const resourceName = "elemento_kaas.test"
	const resourceFormat = `
resource "elemento_kaas" "test" {
  name              = %[1]q
  node_pools = [
    {
      name              = "my-node-pool"
      auto_scale        = %[2]t
      min_nodes         = %[3]d
      max_nodes         = %[4]d
      shared_storage_gb = %[5]d
      specs = {
        cpu_cores    = %[6]d
        memory_gb    = %[7]d
        disk_size_gb = %[8]d
        gpus         = %[9]d
      }
    }
  ]
  nodes_subnet = {
    dhcp_enabled = %[10]t
    cidr         = %[11]q
  }
  update_policy            = %[12]q
  kubernetes_version       = %[13]q
  control_plane_ha_enabled = %[14]t
  region                   = %[15]q
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
				p.nodePoolAutoScale,
				p.nodePoolMinNodes,
				p.nodePoolMaxNodes,
				p.nodePoolSharedStorageGB,
				p.nodePoolCPUCores,
				p.nodePoolMemoryGB,
				p.nodePoolDiskSizeGB,
				p.nodePoolGPUs,
				p.dhcpEnabled,
				p.cidr,
				p.updatePolicy,
				p.kubernetesVersion,
				p.controlPlaneHAEnabled,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "node_pools.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.name", "my-node-pool"),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.auto_scale", strconv.FormatBool(p.nodePoolAutoScale)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.min_nodes", strconv.Itoa(p.nodePoolMinNodes)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.max_nodes", strconv.Itoa(p.nodePoolMaxNodes)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.shared_storage_gb", strconv.Itoa(p.nodePoolSharedStorageGB)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.specs.cpu_cores", strconv.Itoa(p.nodePoolCPUCores)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.specs.memory_gb", strconv.Itoa(p.nodePoolMemoryGB)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.specs.disk_size_gb", strconv.Itoa(p.nodePoolDiskSizeGB)),
				resource.TestCheckResourceAttr(resourceName, "node_pools.0.specs.gpus", strconv.Itoa(p.nodePoolGPUs)),

				resource.TestCheckResourceAttr(resourceName, "nodes_subnet.dhcp_enabled", strconv.FormatBool(p.dhcpEnabled)),
				resource.TestCheckResourceAttr(resourceName, "nodes_subnet.cidr", p.cidr),

				resource.TestCheckResourceAttr(resourceName, "update_policy", p.updatePolicy),
				resource.TestCheckResourceAttr(resourceName, "kubernetes_version", p.kubernetesVersion),
				resource.TestCheckResourceAttr(resourceName, "control_plane_ha_enabled", strconv.FormatBool(p.controlPlaneHAEnabled)),
			}
		},
		[]string{"payment_url", "status", "last_updated"},
	)
}
