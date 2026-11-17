package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPublicGatewayResource(t *testing.T) {
	const resourceName = "elemento_publicgateway.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_privatenetwork" "net" {
  name             = %[1]q
  cidr             = "192.168.2.0/24"
  auto_assign_cidr = false
  region           = "eu-south-1"
}

resource "elemento_publicgateway" "test" {
  name         = %[1]q
  region       = "eu-south-1"
  network_uuid = elemento_privatenetwork.net.uuid
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttrPair(resourceName, "network_uuid", "elemento_privatenetwork.net", "uuid"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
