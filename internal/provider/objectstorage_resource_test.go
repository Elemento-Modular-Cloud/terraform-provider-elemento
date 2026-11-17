package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccObjectStorageResource(t *testing.T) {
	const resourceName = "elemento_objectstorage.test"
	const resourceFormat = `
resource "elemento_objectstorage" "test" {
  name              = %[1]q
  size_tb           = 1
  region            = "%[2]s"
}
`

	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(
				resourceFormat,
				name,
				"eu-south-1",
			)
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
