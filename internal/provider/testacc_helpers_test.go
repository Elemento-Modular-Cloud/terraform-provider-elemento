package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheck(t *testing.T) {
	for _, envVar := range []string{"ELEMENTO_TARGET", "ELEMENTO_USERNAME", "ELEMENTO_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

// testAccCreateImport runs the common create/read + import test used by resources
// whose attributes are all RequiresReplace (so there is no in-place update step).
// config receives a random name; ignore lists attributes the API does not return on read.
func testAccCreateImport(t *testing.T, resourceName string, config func(name string) string, checks func(name string) []resource.TestCheckFunc, ignore []string) {
	rName := acctest.RandomWithPrefix("test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(rName),
				Check:  resource.ComposeAggregateTestCheckFunc(checks(rName)...),
			},
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
