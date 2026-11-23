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

func (p ProviderConfig) ProviderBlock() string {
	return fmt.Sprintf(`
provider "elemento" {
  target = %q
}
`, p.target)
}

type ProviderConfig struct {
	name   string
	target string
	region string
}

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"elemento": providerserver.NewProtocol6WithError(New("test")()),
	}
	targets = map[string]ProviderConfig{
		"aws": {name: "aws", target: "aws-dev-private", region: "eu-south-1"},
		"scaleway": {name: "scaleway", target: "scaleway-dev-private", region: "it-mil"},
	}
	services = map[string][]string{
		"aiserverless": {"aws"},
		"dbaas": {"aws"},
		"filestorage": {"aws"},
		"kaas": {"aws"},
		"loadbalancer": {"aws"},
		"n8n": {"aws"},
		"objectstorage": {"aws"},
		"openclaw": {"aws"},
		"privatenetwork": {"aws"},
		"publicgateway": {"aws"},
		"publicip": {"aws"},
		"queueaas": {"aws"},
		"registry": {"aws"},
		"serverless": {"aws"},
		"spotvm": {"aws"},
		"storage": {"aws"},
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
	for _, envVar := range []string{"ELEMENTO_USERNAME", "ELEMENTO_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

func testAccCreateImport(
	t *testing.T,
	resourceName string,
	resourceType string,
	config func(name string, target ProviderConfig) string,
	checks func(name string, target ProviderConfig) []resource.TestCheckFunc,
	ignore []string,
) {
	providerNames, ok := services[resourceType]
	if !ok || len(providerNames) == 0 {
		t.Fatalf("no target providers registered in services map for %q", resourceType)
	}

	for _, providerName := range providerNames {
		target, ok := targets[providerName]
		if !ok {
			t.Fatalf("unknown target %q referenced by %q", providerName, resourceType)
		}

		t.Run(target.name, func(t *testing.T) {
			rName := acctest.RandomWithPrefix("test")
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: config(rName, target),
						Check:  resource.ComposeAggregateTestCheckFunc(checks(rName, target)...),
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
		})
	}
}

func ptr[T any](s T) *T { return &s }
