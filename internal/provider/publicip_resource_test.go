package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPublicIPResource(t *testing.T) {
	const resourceName = "elemento_publicip.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_publicip" "test" {
  name   = %[1]q
  region = "eu-south-1"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "region", "eu-south-1"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
