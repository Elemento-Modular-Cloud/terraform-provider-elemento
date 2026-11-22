package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type loadBalancerTestParams struct {
	balancerType string
	isPublic     bool
	port         int
}

func TestAccLoadBalancerResource(t *testing.T) {
	paramsByProvider := map[string]loadBalancerTestParams{
		"aws": {
			balancerType: "application",
			isPublic:     false,
			port:         80,
		},
	}

	const resourceType = "loadbalancer"
	const resourceName = "elemento_loadbalancer.test"
	const resourceFormat = `
resource "elemento_loadbalancer" "test" {
  name      = %[1]q
  type      = %[2]q
  is_public = %[3]t
  port      = %[4]d
  region    = %[5]q
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
				p.balancerType,
				p.isPublic,
				p.port,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "type", p.balancerType),
				resource.TestCheckResourceAttr(resourceName, "is_public", strconv.FormatBool(p.isPublic)),
				resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(p.port)),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Computed attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
			}
		},
		[]string{"payment_url", "status", "last_updated"},
	)
}
