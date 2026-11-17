package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccSpotVMResource(t *testing.T) {
	rName := acctest.RandomWithPrefix("test")
	resourceName := "elemento_spotvm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSpotVMPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read testing
			{
				Config: testAccSpotVMResourceConfig(rName, 2, 4096, `[]`, "eu-south-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "vm_name", rName),

					// CPU nested attributes
					resource.TestCheckResourceAttr(resourceName, "cpu.slots", "2"),
					resource.TestCheckResourceAttr(resourceName, "cpu.shared_cores", "true"),
					resource.TestCheckResourceAttr(resourceName, "cpu.arch.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cpu.arch.0", "X86_64"),

					// Memory nested attributes
					resource.TestCheckResourceAttr(resourceName, "mem.capacity_mb", "4096"),

					// OS nested attributes
					resource.TestCheckResourceAttr(resourceName, "os.family", "linux"),
					resource.TestCheckResourceAttr(resourceName, "os.flavour", "ubuntu"),
					resource.TestCheckResourceAttr(resourceName, "os.version", "24.04"),

					// Storage nested attributes
					resource.TestCheckResourceAttr(resourceName, "storage.boot.name", "root"),
					resource.TestCheckResourceAttr(resourceName, "storage.boot.size_gb", "32"),
					resource.TestCheckResourceAttr(resourceName, "storage.data.#", "0"),

					// Auth nested attributes
					resource.TestCheckResourceAttr(resourceName, "auth.username", "elemento"),
					resource.TestCheckResourceAttr(resourceName, "auth.password", "elemento123"),
					resource.TestCheckResourceAttr(resourceName, "auth.ssh_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v"),

					// Top-level attributes
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "region", "eu-south-1"),
					resource.TestCheckResourceAttr(resourceName, "cloudinit_b64", "I2Nsb3VkLWNvbmZpZwpwYWNrYWdlczoKICAtIG5naW54CnJ1bmNtZDoKICAtIFsgc3lzdGVtY3RsLCBlbmFibGUsIG5naW54IF0K"),
				),
			},
			// 2. ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore write-only/sensitive attributes not returned by the API on Read
				ImportStateVerifyIgnore: []string{"auth.password", "cloudinit_b64"},
				// Uncomment if your ImportState uses vm_name instead of an id attribute:
				// ImportStateId: rName,
			},
			// 3. Update and Read testing
			{
				Config: testAccSpotVMResourceConfig(rName, 2, 8192, `["env:test"]`, "eu-south-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "vm_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mem.capacity_mb", "8192"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "env:test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSpotVMPreCheck(t *testing.T) {
	for _, envVar := range []string{"ELEMENTO_TARGET", "ELEMENTO_USERNAME", "ELEMENTO_PASSWORD"} {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

func testAccSpotVMResourceConfig(vmName string, slots, memoryMB int, tagsHCL, region string) string {
	return providerConfig + fmt.Sprintf(`
resource "elemento_spotvm" "test" {
  vm_name = %[1]q
  cpu = {
    slots        = %[2]d
    shared_cores = true
    arch         = ["X86_64"]
  }
  mem = {
    capacity_mb = %[3]d
  }
  os = {
    family  = "linux"
    flavour = "ubuntu"
    version = "24.04"
  }
  storage = {
    boot = {
      name    = "root"
      size_gb = 32
    }
    data = []
  }
  auth = {
    username = "elemento"
    password = "elemento123"
    ssh_key  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v"
  }
  tags          = %[4]s
  region        = %[5]q
  cloudinit_b64 = "I2Nsb3VkLWNvbmZpZwpwYWNrYWdlczoKICAtIG5naW54CnJ1bmNtZDoKICAtIFsgc3lzdGVtY3RsLCBlbmFibGUsIG5naW54IF0K"
}
`,
		vmName,
		slots,
		memoryMB,
		tagsHCL,
		region,
	)
}
