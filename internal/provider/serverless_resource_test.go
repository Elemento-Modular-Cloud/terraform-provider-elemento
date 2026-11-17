package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerlessResource(t *testing.T) {
	const resourceName = "elemento_serverless.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_serverless" "test" {
  name      = %[1]q
  region    = "eu-south-1"
  runtime   = "python3.13"
  handler   = "main.handler"
  memory    = 512
  timeout   = 30
  is_public = true
  environment = {
    STAGE = "test"
  }
  code_zip_b64 = "UEsDBBQAAAAIAHJRFV2KSnlhqQAAAAgBAAAHABwAbWFpbi5weVVUCQADWCSIalgkiGp1eAsAAQQAAAAABAAAAABljrEOwjAMRPd+xZGJShFCjF1Z2PkCNzEkUokhdQsV4t8pAQkhvFi+s9/Z8wGBku94ySMntXCSlG9aNxXmMsaUXswGPjotGxRTTEdkvgzcKzwpYXliDeItzqTBIjB5zr1FK36yYHWrurA+Cf+0+Z4KiVoZFBoYMY3iSKOkn3cy65AT7mUoRq+kQ78Vz6bBZr22X+uVP4tmx10nOGQ5Ye+o4ytNizfuUT0BUEsBAh4DFAAAAAgAclEVXYpKeWGpAAAACAEAAAcAGAAAAAAAAQAAAKSBAAAAAG1haW4ucHlVVAUAA1gkiGp1eAsAAQQAAAAABAAAAABQSwUGAAAAAAEAAQBNAAAA6gAAAAAA"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "runtime", "python3.13"),
				resource.TestCheckResourceAttr(resourceName, "handler", "main.handler"),
				resource.TestCheckResourceAttr(resourceName, "memory", "512"),
			}
		},
		[]string{"code_zip_b64", "payment_url", "status"},
	)
}
