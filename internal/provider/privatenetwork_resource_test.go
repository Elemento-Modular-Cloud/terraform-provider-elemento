package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrivateNetworkResource(t *testing.T) {
	const resourceName = "elemento_privatenetwork.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_privatenetwork" "test" {
  name             = %[1]q
  cidr             = "192.168.2.0/24"
  auto_assign_cidr = false
  region           = "eu-south-1"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "cidr", "192.168.2.0/24"),
				resource.TestCheckResourceAttr(resourceName, "auto_assign_cidr", "false"),
			}
		},
		[]string{"payment_url", "status", "auto_assign_cidr"},
	)
}
