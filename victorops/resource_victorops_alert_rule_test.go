package victorops

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAlertRuleCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	tfResourceName := "victorops_alert_rule.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		CheckDestroy: testAccAlertRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertRuleConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccAlertRuleExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "alert_field", "host_name"),
					resource.TestCheckResourceAttr(tfResourceName, "alert_value_match", "test-*"),
					resource.TestCheckResourceAttr(tfResourceName, "match_type", "WILDCARD"),
					resource.TestCheckResourceAttr(tfResourceName, "stop_flag", "false"),
				),
			},
		},
	})
}

func testAccAlertRuleConfig() string {
	return `
resource "victorops_alert_rule" "test" {
  alert_field       = "host_name"
  alert_value_match = "test-*"
  match_type        = "WILDCARD"
  routing_key       = "everyone"
  stop_flag         = false
  notes             = "Test alert rule created by Terraform acceptance test"
}
`
}

func testAccAlertRuleExists(resourceName string) resource.TestCheckFunc {
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

func testAccAlertRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_alert_rule" {
			continue
		}
		// Alert rule should be deleted
	}
	return nil
}
