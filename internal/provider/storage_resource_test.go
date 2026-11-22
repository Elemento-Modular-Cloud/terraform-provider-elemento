package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type storageTestParams struct {
	private bool
	readonly bool
	shareable bool
	bootable bool
	clonable bool
	alg string
	format string
	bus string
	size int
	priority int
	tags string
}

func TestAccStorageResource(t *testing.T) {
	paramsByProvider := map[string]storageTestParams{
		"aws": {
			private: false,
			readonly: false,
			shareable: false,
			bootable: true,
			clonable: true,
			alg: "cp",
			format: "qcow2",
			bus: "virtio",
			size: 32,
			priority: 100,
			tags: `["test:test"]`,
		},
	}

	const resourceType = "storage"
	const resourceName = "elemento_storage.test"
	const resourceFormat = `
resource "elemento_storage" "test" {
  name      = %[1]q
  private   = %[2]t
  readonly  = %[3]t
  shareable = %[4]t
  bootable  = %[5]t
  clonable  = %[6]t
  alg       = %[7]q
  format    = %[8]q
  bus       = %[9]q
  size      = %[10]d
  priority  = %[11]d
  tags      = %[12]s
  region    = %[13]q
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
				p.private,
				p.readonly,
				p.shareable,
				p.bootable,
				p.clonable,
				p.alg,
				p.format,
				p.bus,
				p.size,
				p.priority,
				p.tags,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "private", strconv.FormatBool(p.private)),
				resource.TestCheckResourceAttr(resourceName, "readonly", strconv.FormatBool(p.readonly)),
				resource.TestCheckResourceAttr(resourceName, "shareable", strconv.FormatBool(p.shareable)),
				resource.TestCheckResourceAttr(resourceName, "bootable", strconv.FormatBool(p.bootable)),
				resource.TestCheckResourceAttr(resourceName, "clonable", strconv.FormatBool(p.clonable)),
				resource.TestCheckResourceAttr(resourceName, "alg", p.alg),
				resource.TestCheckResourceAttr(resourceName, "format", p.format),
				resource.TestCheckResourceAttr(resourceName, "bus", p.bus),
				resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(p.size)),
				resource.TestCheckResourceAttr(resourceName, "priority", strconv.Itoa(p.priority)),
				resource.TestCheckResourceAttr(resourceName, "tags", p.tags),
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
