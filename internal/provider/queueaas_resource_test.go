package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type queueAASTestParams struct {
	cidr string
	autoAssignCidr bool
	nodeNumber int
	nodeSize string
	sizeGB int
	username string
	password string
	kafkaVersion string
	isHA bool
	isPublic bool
}

func TestAccQueueAASResource(t *testing.T) {
	paramsByProvider := map[string]queueAASTestParams{
		"aws": {
			cidr: "192.168.2.0/24",
			autoAssignCidr: false,
			nodeNumber: 3,
			nodeSize: "M",
			sizeGB: 10,
			username: "user",
			password: "1234515438u_sdf!",
			kafkaVersion: "3.9.1",
			isHA: true,
			isPublic: false,
		},
	}

	const resourceType = "queueaas"
	const resourceName = "elemento_queueaas.test"
	const resourceFormat = `
resource "elemento_privatenetwork" "net" {
  name             = "net"
  cidr             = %[2]q
  auto_assign_cidr = %[3]t
  region           = %[12]q
}

resource "elemento_queueaas" "test" {
  name          = %[1]q
  node_number   = %[4]d
  node_size     = %[5]q
  size_gb       = %[6]d
  username      = %[7]q
  password      = %[8]q
  kafka_version = %[9]q
  is_ha         = %[10]t
  is_public     = %[11]t
  network_uuid  = elemento_privatenetwork.net.uuid
  region        = %[12]q
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
				p.cidr,
				p.autoAssignCidr,
				p.nodeNumber,
				p.nodeSize,
				p.sizeGB,
				p.username,
				p.password,
				p.kafkaVersion,
				p.isHA,
				p.isPublic,
				target.region,
			)
		},
		func(name string, target ProviderConfig) []resource.TestCheckFunc {
			p := paramsByProvider[target.name]

			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "name", name),
				resource.TestCheckResourceAttr(resourceName, "node_number", strconv.Itoa(p.nodeNumber)),
				resource.TestCheckResourceAttr(resourceName, "node_size", p.nodeSize),
				resource.TestCheckResourceAttr(resourceName, "size_gb", strconv.Itoa(p.sizeGB)),
				resource.TestCheckResourceAttr(resourceName, "username", p.username),
				resource.TestCheckResourceAttr(resourceName, "password", p.password),
				resource.TestCheckResourceAttr(resourceName, "kafka_version", p.kafkaVersion),
				resource.TestCheckResourceAttr(resourceName, "is_ha", strconv.FormatBool(p.isHA)),
				resource.TestCheckResourceAttr(resourceName, "is_public", strconv.FormatBool(p.isPublic)),
				resource.TestCheckResourceAttr(resourceName, "region", target.region),

				// Public resource attributes
				resource.TestCheckNoResourceAttr(resourceName, "payment_url"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_frequency"),
				resource.TestCheckNoResourceAttr(resourceName, "billing_uuid"),
			}
		},
		[]string{"password", "payment_url", "status", "last_updated"},
	)
}
