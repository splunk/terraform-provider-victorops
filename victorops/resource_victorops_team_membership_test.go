package victorops

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/victorops/go-victorops/victorops"
	"os"
	"strings"
	"testing"
)

type MembershipData struct {
	User        victorops.User
	TeamName    string
	Replacement string
}

func TestCreate_TeamMembership(t *testing.T) {
	tfMembershipResourceName := "victorops_team_membership.test_membership"
	membership := createNewMembershipModel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: createMembershipResource(membership),
				Check: resource.ComposeTestCheckFunc(
					testAccTeamMembershipExists(tfMembershipResourceName),
					resource.TestCheckResourceAttr(tfMembershipResourceName, "user_name", membership.User.Username),
				),
			},
		},
	})
}

func createNewMembershipModel() MembershipData {
	replacementUser := os.Getenv("VO_REPLACEMENT_USERNAME")

	return MembershipData{
		User: victorops.User{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Username:  strings.ToLower(faker.Username()),
			Email:     faker.Email(),
			Admin:     true,
		},
		TeamName:    faker.Word(),
		Replacement: replacementUser,
	}
}

func createMembershipResource(md MembershipData) string {
	return getTestTemplate("test_membership.tf", md)
}

func testAccTeamMembershipExists(resource string) resource.TestCheckFunc {
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

func testMembershipDestroy(s *terraform.State) error {
	return nil
}
