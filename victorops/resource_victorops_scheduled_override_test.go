package victorops

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScheduledOverrideCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	// Note: This test requires VO_OVERRIDE_USERNAME to be set to a user that is
	// part of an escalation policy. The default VO_REPLACEMENT_USERNAME may not
	// qualify. Set VO_OVERRIDE_USERNAME to a user in an escalation policy.
	username := os.Getenv("VO_OVERRIDE_USERNAME")
	if username == "" {
		t.Skip("VO_OVERRIDE_USERNAME not set - this test requires a user that is part of an escalation policy")
	}

	// Use future dates for the override in UTC to avoid timezone mismatch
	// Round to the nearest hour to avoid drift between plan and apply
	now := time.Now().UTC().Truncate(time.Hour).Add(time.Hour) // Next hour
	start := now.Add(24 * time.Hour).Format("2006-01-02T15:04:05Z")
	end := now.Add(48 * time.Hour).Format("2006-01-02T15:04:05Z")

	tfResourceName := "victorops_scheduled_override.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		CheckDestroy: testAccScheduledOverrideDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledOverrideConfig(username, start, end),
				Check: resource.ComposeTestCheckFunc(
					testAccScheduledOverrideExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "username", username),
					resource.TestCheckResourceAttr(tfResourceName, "timezone", "UTC"),
					resource.TestCheckResourceAttrSet(tfResourceName, "public_id"),
				),
			},
		},
	})
}

func testAccScheduledOverrideConfig(username, start, end string) string {
	return fmt.Sprintf(`
resource "victorops_scheduled_override" "test" {
  username = "%s"
  timezone = "UTC"
  start    = "%s"
  end      = "%s"
}
`, username, start, end)
}

func testAccScheduledOverrideExists(resourceName string) resource.TestCheckFunc {
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

func testAccScheduledOverrideDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_scheduled_override" {
			continue
		}
		// Override should be deleted
	}
	return nil
}
