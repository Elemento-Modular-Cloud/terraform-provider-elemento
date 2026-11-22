package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type privateNetworkTestParams struct {
	cidr           string
	autoAssignCidr bool
}

func TestAccPrivateNetworkResource(t *testing.T) {
	paramsByProvider := map[string]privateNetworkTestParams{
		"aws": {
			cidr:           "192.168.2.0/24",
			autoAssignCidr: false,
		},
	}

	const resourceType = "privatenetwork"
	const resourceName = "elemento_privatenetwork.test"
	const resourceFormat = `
resource "elemento_privatenetwork" "test" {
  name             = %[1]q
  cidr             = %[2]q
  auto_assign_cidr = %[3]t
  region           = %[4]q
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
				p.cidr,
				p.autoAssignCidr,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "cidr", p.cidr),
				resource.TestCheckResourceAttr(resourceName, "auto_assign_cidr", strconv.FormatBool(p.autoAssignCidr)),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
			}
		},
		[]string{"payment_url", "status", "auto_assign_cidr", "last_updated"},
	)
}
