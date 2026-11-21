package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type fileStorageTestParams struct {
	sizeGB   int
	isPublic bool
	region   *string
}

func TestAccFileStorageResource(t *testing.T) {
	paramsByProvider := map[string]fileStorageTestParams{
		"aws": {
			sizeGB:   10,
			isPublic: false,
		},
		"scaleway": {
			sizeGB:   30,
			isPublic: false,
			region:   ptr("fr-par"),
		},
	}

	const resourceType = "filestorage"
	const resourceName = "elemento_filestorage.test"
	const resourceFormat = `
resource "elemento_filestorage" "test" {
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

			region := target.region
			if p.region != nil {
				region = *p.region
			}

			return target.ProviderBlock() + fmt.Sprintf(
				resourceFormat,
				name,
				p.sizeGB,
				p.isPublic,
				region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "size_gb", strconv.Itoa(p.sizeGB)),
				resource.TestCheckResourceAttr(resourceName, "is_public", strconv.FormatBool(p.isPublic)),

				// Computed attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
			}
		},
		[]string{"payment_url", "status", "last_updated", "is_public"},
	)
}
