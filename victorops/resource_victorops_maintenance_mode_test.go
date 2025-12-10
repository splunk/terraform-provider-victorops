package victorops

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMaintenanceModeCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	tfResourceName := "victorops_maintenance_mode.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		CheckDestroy: testAccMaintenanceModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceModeConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccMaintenanceModeExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "purpose", "Terraform acceptance test"),
					resource.TestCheckResourceAttrSet(tfResourceName, "instance_id"),
				),
			},
		},
	})
}

func testAccMaintenanceModeConfig() string {
	// Use global maintenance mode (empty routing_keys) for testing
	return `
resource "victorops_maintenance_mode" "test" {
  purpose = "Terraform acceptance test"
}
`
}

func testAccMaintenanceModeExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		return nil
	}
}

func testAccMaintenanceModeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_maintenance_mode" {
			continue
		}
		// Maintenance mode should be ended
	}
	return nil
}
