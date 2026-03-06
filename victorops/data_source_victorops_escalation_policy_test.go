package victorops

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceEscalationPolicyRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	teamName := "TF-DataSource-Policy-Team"
	policyName := "TF-DataSource-Policy"
	tfDataSourceName := "data.victorops_escalation_policy.test"
	tfResourceName := "victorops_escalation_policy.ds_test_policy"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEscalationPolicyConfig(teamName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceEscalationPolicyExists(tfResourceName),
					testAccDataSourceEscalationPolicyExists(tfDataSourceName),
					resource.TestCheckResourceAttr(tfDataSourceName, "name", policyName),
					resource.TestCheckResourceAttrSet(tfDataSourceName, "slug"),
					resource.TestCheckResourceAttrSet(tfDataSourceName, "team_slug"),
				),
			},
		},
	})
}

func testAccDataSourceEscalationPolicyConfig(teamName, policyName string) string {
	return fmt.Sprintf(`
resource "victorops_team" "ds_ep_team" {
  name = "%s"
}

resource "victorops_escalation_policy" "ds_test_policy" {
  name    = "%s"
  team_id = victorops_team.ds_ep_team.id
  step {
    timeout = 0
    entries = []
  }
}

data "victorops_escalation_policy" "test" {
  slug = victorops_escalation_policy.ds_test_policy.id
}
`, teamName, policyName)
}

func testAccDataSourceEscalationPolicyExists(resourceOrDataSource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceOrDataSource]
		if !ok {
			return fmt.Errorf("resource/data source not found: %s", resourceOrDataSource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", resourceOrDataSource)
		}
		return nil
	}
}
