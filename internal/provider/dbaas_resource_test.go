package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDBaaSResource(t *testing.T) {
	rName := acctest.RandomWithPrefix("test")
	resourceName := "elemento_dbaas.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDBaaSPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read testing
			{
				Config: testAccDBaaSResourceConfig(rName, 50, "12:00:00", 1, "eu-south-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "postgres"),
					resource.TestCheckResourceAttr(resourceName, "disk_size_gb", "50"),
					resource.TestCheckResourceAttr(resourceName, "backup_time", "12:00:00"),
					resource.TestCheckResourceAttr(resourceName, "password", "mysecretpassword"),
					resource.TestCheckResourceAttr(resourceName, "nodes_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "region", "eu-south-1"),
					resource.TestCheckResourceAttr(resourceName, "billing_frequency", "day"),

					// Computed attributes
					resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				),
			},
			// 2. ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore write-only/sensitive or creation-only attributes not returned by the API on Read
				ImportStateVerifyIgnore: []string{"password", "payment_url", "status"},
				// Uncomment if your ImportState uses name instead of an id attribute:
				// ImportStateId: rName,
				ImportStateIdFunc: AttrImportStateIdFunc(resourceName, "uuid"),
				ImportStateVerifyIdentifierAttribute: "uuid",
			},
			// 3. Update and Read testing
			{
				Config: testAccDBaaSResourceConfig(rName, 100, "14:00:00", 1, "fi-hel2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "disk_size_gb", "100"),
					resource.TestCheckResourceAttr(resourceName, "backup_time", "14:00:00"),
					resource.TestCheckResourceAttr(resourceName, "nodes_number", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDBaaSPreCheck(t *testing.T) {
	for _, envVar := range []string{"ELEMENTO_TARGET", "ELEMENTO_USERNAME", "ELEMENTO_PASSWORD"} {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

func testAccDBaaSResourceConfig(name string, diskSizeGB int, backupTime string, nodesNumber int, region string) string {
	return providerConfig + fmt.Sprintf(`
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
`,
		name,
		diskSizeGB,
		backupTime,
		nodesNumber,
		region,
	)
}
