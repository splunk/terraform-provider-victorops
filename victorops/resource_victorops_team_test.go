package victorops

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"regexp"
	"testing"
)

func TestAccTeamCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	teamName := "DevOps"
	tfResourceName := "victorops_team.test_team"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: createTeamResource(teamName),
				Check: resource.ComposeTestCheckFunc(
					testAccTeamExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "name", teamName),
				),
			},
		},
	})
}

func createTeamResource(s string) string {
	return getTestTemplate("test_team.tf", s)
}

func testAccTeamExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		teamSlug := rs.Primary.ID
		apiClient := testAccProvider.Meta().(Config).VictorOpsClient
		_, _, err := apiClient.GetTeam(teamSlug)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testTeamDestroy(s *terraform.State) error {

	apiClient := testAccProvider.Meta().(Config).VictorOpsClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_team" {
			continue
		}

		_, response, _ := apiClient.GetTeam(rs.Primary.ID)
		notFoundErr := "{\"error\":\"No team 'team-.+' found\"}"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(response.ResponseBody)) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, response.ResponseBody)
		}
	}

	return nil
}
