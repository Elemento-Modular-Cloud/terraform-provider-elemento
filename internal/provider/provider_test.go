package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
    // providerConfig is a shared configuration to combine with the actual
    // test configuration so the HashiCups client is properly configured.
    // It is also possible to use the HASHICUPS_ environment variables instead,
    // such as updating the Makefile and running the testing through that tool.
    providerConfig = `
provider "elemento" {
  target = "aws-dev-private"
}
`
)

var (
    // testAccProtoV6ProviderFactories are used to instantiate a provider during
    // acceptance testing. The factory function will be invoked for every Terraform
    // CLI command executed to create a provider server to which the CLI can
    // reattach.
    testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
        "elemento": providerserver.NewProtocol6WithError(New("test")()),
    }
)

func AttrImportStateIdFunc(resourceName, attrName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		val, ok := rs.Primary.Attributes[attrName]
		if !ok || val == "" {
			return "", fmt.Errorf("attribute %q not set on %s", attrName, resourceName)
		}

		return val, nil
	}
}

func testAccPreCheck(t *testing.T) {
	for _, envVar := range []string{"ELEMENTO_TARGET", "ELEMENTO_USERNAME", "ELEMENTO_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

func testAccCreateImport(t *testing.T, resourceName string, config func(name string) string, checks func(name string) []resource.TestCheckFunc, ignore []string) {
	rName := acctest.RandomWithPrefix("test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: config(rName),
				Check:  resource.ComposeAggregateTestCheckFunc(checks(rName)...),
			},
			// ImportState testing
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              ignore,
				ImportStateIdFunc:                    AttrImportStateIdFunc(resourceName, "uuid"),
				ImportStateVerifyIdentifierAttribute: "uuid",
			},
		},
	})
}
