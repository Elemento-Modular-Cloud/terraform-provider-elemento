package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dbaasTestParams struct {
	engine      string
	diskSizeGB  int
	backupTime  string
	nodesNumber int
}

func TestAccDBaaSResource(t *testing.T) {
	paramsByProvider := map[string]dbaasTestParams{
		"scaleway": {
			engine:      "postgres",
			diskSizeGB:  50,
			backupTime:  "12:00:00",
			nodesNumber: 1,
		},
		"aws": {
			engine:      "postgres",
			diskSizeGB:  50,
			backupTime:  "12:00:00",
			nodesNumber: 1,
		},
	}

	const resourceType = "dbaas"
	const resourceName = "elemento_dbaas.test"
	const resourceFormat = `
resource "elemento_dbaas" "test" {
  name              = %[1]q
  engine            = %[2]q
  disk_size_gb      = %[3]d
  backup_time       = %[4]q
  password          = %[5]q
  nodes_number      = %[6]d
  region            = %[7]q
}
`

	password := acctest.RandomWithPrefix("Password123_!")

	testAccCreateImport(
		t,
		resourceName,
		resourceType,
		func(name string, target ProviderConfig) string {
			p := paramsByProvider[target.name]

			return target.ProviderBlock() + fmt.Sprintf(
				resourceFormat,
				name,
				p.engine,
				p.diskSizeGB,
				p.backupTime,
				password,
				p.nodesNumber,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "engine", p.engine),
				resource.TestCheckResourceAttr(resourceName, "disk_size_gb", strconv.Itoa(p.diskSizeGB)),
				resource.TestCheckResourceAttr(resourceName, "backup_time", p.backupTime),
				resource.TestCheckResourceAttr(resourceName, "password", password),
				resource.TestCheckResourceAttr(resourceName, "nodes_number", strconv.Itoa(p.nodesNumber)),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Computed attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
			}
		},
		[]string{"password", "payment_url", "status", "last_updated"},
	)
}
