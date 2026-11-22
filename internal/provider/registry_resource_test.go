package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type registryTestParams struct {
	sizeGB   int
	isPublic bool
}

func TestAccRegistryResource(t *testing.T) {
	paramsByProvider := map[string]registryTestParams{
		"aws": {
			sizeGB:   10,
			isPublic: false,
		},
	}

	const resourceType = "registry"
	const resourceName = "elemento_registry.test"
	const resourceFormat = `
resource "elemento_registry" "test" {
  name      = %[1]q
  size_gb   = %[2]d
  is_public = %[3]t
  region    = %[4]q
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
				p.sizeGB,
				p.isPublic,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "size_gb", strconv.Itoa(p.sizeGB)),
				resource.TestCheckResourceAttr(resourceName, "is_public", strconv.FormatBool(p.isPublic)),
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
