package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type aiServerlessTestParams struct {
	modelName string
	enableStreaming bool
}

func TestAccAIServerlessResource(t *testing.T) {
	paramsByProvider := map[string]aiServerlessTestParams{
		"aws": {
			modelName: "claude-sonnet-5",
			enableStreaming: true,
		},
	}

	const resourceType = "aiserverless"
	const resourceName = "elemento_aiserverless.test"
	const resourceFormat = `
resource "elemento_aiserverless" "test" {
  name             = %[1]q
  model_name       = %[2]q
  enable_streaming = %[3]t
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
				p.modelName,
				p.enableStreaming,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "model_name", p.modelName),
				resource.TestCheckResourceAttr(resourceName, "enable_streaming", strconv.FormatBool(p.enableStreaming)),
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
