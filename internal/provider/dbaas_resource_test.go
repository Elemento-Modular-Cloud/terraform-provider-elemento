package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDBaaSResource(t *testing.T) {
	const resourceName = "elemento_dbaas.test"
	const resourceFormat = `
resource "elemento_dbaas" "test" {
  name              = %[1]q
  engine            = "postgres"
  disk_size_gb      = %[2]d
  backup_time       = %[3]q
  password          = "mysecretpassword"
  nodes_number      = %[4]d
  region            = %[5]q
  billing_frequency = "day"
}
`

	testAccCreateImport(
		t,
		resourceName,
		func(name string) string {
			return providerConfig + fmt.Sprintf(
				resourceFormat,
				name,
				50,
				"12:00:00",
				1,
				"eu-south-1",
			)
		},
		func(name string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "engine", "postgres"),
				resource.TestCheckResourceAttr(resourceName, "disk_size_gb", "50"),
				resource.TestCheckResourceAttr(resourceName, "backup_time", "12:00:00"),
				resource.TestCheckResourceAttr(resourceName, "password", "mysecretpassword"),
				resource.TestCheckResourceAttr(resourceName, "nodes_number", "1"),
				resource.TestCheckResourceAttr(resourceName, "region", "eu-south-1"),
				resource.TestCheckResourceAttr(resourceName, "billing_frequency", "day"),

				// Computed attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
			}
		},
		[]string{"password", "payment_url", "status", "last_updated"},
	)
}
