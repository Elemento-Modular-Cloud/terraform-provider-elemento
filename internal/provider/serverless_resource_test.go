package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type serverlessTestParams struct {
	runtime    string
	handler    string
	memory     int
	timeout    int
	isPublic   bool
	stage      string
	codeZipB64 string
}

func TestAccServerlessResource(t *testing.T) {
	paramsByProvider := map[string]serverlessTestParams{
		"aws": {
			runtime:    "python3.13",
			handler:    "main.handler",
			memory:     512,
			timeout:    30,
			isPublic:   true,
			stage:      "test",
			codeZipB64: "UEsDBBQAAAAIAHJRFV2KSnlhqQAAAAgBAAAHABwAbWFpbi5weVVUCQADWCSIalgkiGp1eAsAAQQAAAAABAAAAABljrEOwjAMRPd+xZGJShFCjF1Z2PkCNzEkUokhdQsV4t8pAQkhvFi+s9/Z8wGBku94ySMntXCSlG9aNxXmMsaUXswGPjotGxRTTEdkvgzcKzwpYXliDeItzqTBIjB5zr1FK36yYHWrurA+Cf+0+Z4KiVoZFBoYMY3iSKOkn3cy65AT7mUoRq+kQ78Vz6bBZr22X+uVP4tmx10nOGQ5Ye+o4ytNizfuUT0BUEsBAh4DFAAAAAgAclEVXYpKeWGpAAAACAEAAAcAGAAAAAAAAQAAAKSBAAAAAG1haW4ucHlVVAUAA1gkiGp1eAsAAQQAAAAABAAAAABQSwUGAAAAAAEAAQBNAAAA6gAAAAAA",
		},
	}

	const resourceType = "serverless"
	const resourceName = "elemento_serverless.test"
	const resourceFormat = `
resource "elemento_serverless" "test" {
  name        = %[1]q
  runtime     = %[2]q
  handler     = %[3]q
  memory      = %[4]d
  timeout     = %[5]d
  is_public   = %[6]t
  environment = {
    STAGE = %[7]q
  }
  code_zip_b64 = %[8]q
  region       = %[9]q
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
				p.runtime,
				p.handler,
				p.memory,
				p.timeout,
				p.isPublic,
				p.stage,
				p.codeZipB64,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "runtime", p.runtime),
				resource.TestCheckResourceAttr(resourceName, "handler", p.handler),
				resource.TestCheckResourceAttr(resourceName, "memory", strconv.Itoa(p.memory)),
				resource.TestCheckResourceAttr(resourceName, "timeout", strconv.Itoa(p.timeout)),
				resource.TestCheckResourceAttr(resourceName, "is_public", strconv.FormatBool(p.isPublic)),
				resource.TestCheckResourceAttr(resourceName, "environment.STAGE", p.stage),
				resource.TestCheckResourceAttr(resourceName, "code_zip_b64", p.codeZipB64),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
			}
		},
		[]string{"code_zip_b64", "payment_url", "status", "last_updated", "environment"},
	)
}
