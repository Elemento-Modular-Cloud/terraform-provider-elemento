package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerResource(t *testing.T) {
	const resourceName = "elemento_loadbalancer.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_loadbalancer" "test" {
  name      = %[1]q
  type      = "application"
  is_public = false
  port      = 80
  region    = "eu-south-1"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "type", "application"),
				resource.TestCheckResourceAttr(resourceName, "port", "80"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
