package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStorageResource(t *testing.T) {
	const resourceName = "elemento_storage.test"
	testAccCreateImport(t, resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(`
resource "elemento_storage" "test" {
  name              = %[1]q
  private           = false
  readonly          = false
  shareable         = false
  bootable          = true
  clonable          = true
  alg               = "cp"
  format            = "qcow2"
  bus               = "virtio"
  size              = 32
  priority          = 100
  tags              = ["test:test"]
  region            = "italynorth"
  billing_frequency = "month"
}
`, name)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "size", "32"),
				resource.TestCheckResourceAttr(resourceName, "format", "qcow2"),
				resource.TestCheckResourceAttr(resourceName, "billing_frequency", "month"),
			}
		},
		[]string{"payment_url", "status"},
	)
}
