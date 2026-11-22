package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPublicIPResource(t *testing.T) {
	const resourceType = "publicip"
	const resourceName = "elemento_publicip.test"
	const resourceFormat = `
resource "elemento_publicip" "test" {
  name   = %[1]q
  region = "eu-south-1"
}
`

	testAccCreateImport(
		t,
		resourceName,
		resourceType,
		func(name string, target ProviderConfig) string {
			return target.ProviderBlock() + fmt.Sprintf(
				resourceFormat,
				name,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "region", "eu-south-1"),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
			}
		},
		[]string{"payment_url", "status", "last_updated"},
	)
}
