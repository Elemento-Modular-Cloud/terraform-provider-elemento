package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegistryResource(t *testing.T) {
	const resourceName = "elemento_registry.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_registry" "test" {
  name      = %[1]q
  size_gb   = 10
  is_public = false
  region    = "eu-south-1"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "size_gb", "10"),
				resource.TestCheckResourceAttr(resourceName, "is_public", "false"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
