package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQueueAASResource(t *testing.T) {
	const resourceName = "elemento_queueaas.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_privatenetwork" "net" {
  name             = %[1]q
  cidr             = "192.168.2.0/24"
  auto_assign_cidr = false
  region           = "eu-south-1"
}

resource "elemento_queueaas" "test" {
  name          = %[1]q
  region        = "eu-south-1"
  node_number   = 3
  node_size     = "M"
  size_gb       = 10
  username      = "user"
  password      = "1234515438u_sdf!"
  kafka_version = "3.9.1"
  is_ha         = true
  is_public     = false
  network_uuid  = elemento_privatenetwork.net.vpc_id
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "node_number", "3"),
				resource.TestCheckResourceAttr(resourceName, "kafka_version", "3.9.1"),
			}
		},
		[]string{"password", "payment_url", "status"},
	)
}
