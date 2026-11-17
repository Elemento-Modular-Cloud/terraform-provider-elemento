package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccObjectStorageResource(t *testing.T) {
	const resourceName = "elemento_objectstorage.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_objectstorage" "test" {
  name              = %[1]q
  size_tb           = 1
  region            = "italynorth"
  billing_frequency = "day"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "size_tb", "1"),
				resource.TestCheckResourceAttr(resourceName, "billing_frequency", "day"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
