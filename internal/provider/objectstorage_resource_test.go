package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type objectStorageTestParams struct {
	sizeTB int
}

func TestAccObjectStorageResource(t *testing.T) {
	paramsByProvider := map[string]objectStorageTestParams{
		"aws": {
			sizeTB: 1,
		},
	}

	const resourceType = "objectstorage"
	const resourceName = "elemento_objectstorage.test"
	const resourceFormat = `
resource "elemento_objectstorage" "test" {
  name              = %[1]q
  size_tb           = %[2]d
  region            = %[3]q
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
				p.sizeTB,
				"eu-south-1",
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "size_tb", strconv.Itoa(p.sizeTB)),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
			}
		},
		[]string{"payment_url", "status", "last_updated", "size_tb"},
	)
}
