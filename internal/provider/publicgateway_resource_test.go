package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type publicGatewayTestParams struct {
  cidr             string
  autoAssignCidr   bool
}

func TestAccPublicGatewayResource(t *testing.T) {
	paramsByProvider := map[string]publicGatewayTestParams{
		"aws": {
			cidr: "192.168.2.0/24",
			autoAssignCidr: false,
		},
	}

	const resourceType = "publicgateway"
	const resourceName = "elemento_publicgateway.test"
	const resourceFormat = `
resource "elemento_privatenetwork" "net" {
  name             = "test-network"
  cidr             = %[2]q
  auto_assign_cidr = %[3]t
  region           = %[4]q
}

resource "elemento_publicgateway" "test" {
  name         = %[1]q
  region       = %[4]q
  network_uuid = elemento_privatenetwork.net.uuid
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
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttrPair(resourceName, "network_uuid", "elemento_privatenetwork.net", "uuid"),
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
