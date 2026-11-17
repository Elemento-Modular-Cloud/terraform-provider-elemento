package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAIServerlessResource(t *testing.T) {
	const resourceName = "elemento_aiserverless.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_aiserverless" "test" {
  name             = %[1]q
  region           = "eu-south-1"
  model_name       = "claude-sonnet-5"
  enable_streaming = true
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "model_name", "claude-sonnet-5"),
				resource.TestCheckResourceAttr(resourceName, "enable_streaming", "true"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
