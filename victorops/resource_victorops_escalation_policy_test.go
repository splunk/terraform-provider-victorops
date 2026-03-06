package victorops

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccEscalationPolicyCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	teamName := "TF-Acc-EP-Team"
	policyName := "TF-Acc-EP-Policy"
	tfPolicyName := "victorops_escalation_policy.test_ep"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccEscalationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEscalationPolicyConfig(teamName, policyName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccEscalationPolicyExists(tfPolicyName),
					resource.TestCheckResourceAttr(tfPolicyName, "name", policyName),
					resource.TestCheckResourceAttr(tfPolicyName, "step.#", "1"),
					resource.TestCheckResourceAttr(tfPolicyName, "step.0.timeout", "0"),
				),
			},
		},
	})
}

func TestAccEscalationPolicyImport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	teamName := "TF-Acc-EP-Import-Team"
	policyName := "TF-Acc-EP-Import-Policy"
	tfPolicyName := "victorops_escalation_policy.test_ep"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccEscalationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEscalationPolicyConfig(teamName, policyName, 0),
			},
			// Verify that terraform import works using the policy slug (its ID).
			{
				ResourceName:      tfPolicyName,
				ImportState:       true,
				ImportStateVerify: true,
				// The step entries list is not fully read back from the API,
				// so we skip verifying it during import.
				ImportStateVerifyIgnore: []string{"step"},
			},
		},
	})
}

// testAccEscalationPolicyConfig generates HCL for a team + escalation policy with an empty step list.
// This avoids requiring a real rotation group slug in the test environment.
func testAccEscalationPolicyConfig(teamName, policyName string, stepTimeout int) string {
	return fmt.Sprintf(`
resource "victorops_team" "test_ep_team" {
  name = "%s"
}

resource "victorops_escalation_policy" "test_ep" {
  name    = "%s"
  team_id = victorops_team.test_ep_team.id
  step {
    timeout = %d
    entries = []
  }
}
`, teamName, policyName, stepTimeout)
}

func testAccEscalationPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", resourceName)
		}

		apiClient := testAccProvider.Meta().(Config).VictorOpsClient
		_, details, err := apiClient.GetEscalationPolicy(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching escalation policy %s: %s", rs.Primary.ID, err)
		}
		if details.StatusCode != 200 {
			return fmt.Errorf("escalation policy %s not found (%d)", rs.Primary.ID, details.StatusCode)
		}
		return nil
	}
}

func testAccEscalationPolicyDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(Config).VictorOpsClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_escalation_policy" {
			continue
		}

		_, response, _ := apiClient.GetEscalationPolicy(rs.Primary.ID)
		notFoundPattern := regexp.MustCompile(`(?i)(not found|no.*policy|404)`)
		if !notFoundPattern.MatchString(response.ResponseBody) && response.StatusCode != 404 {
			return fmt.Errorf("escalation policy %s still exists (status %d)", rs.Primary.ID, response.StatusCode)
		}
	}

	return nil
}
