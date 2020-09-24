package victorops

import (
	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/victorops/go-victorops/victorops"
	"os"
	"strings"
	"testing"
)


type MembershipData struct {
	User victorops.User
	TeamName string
	Replacement string
}

func TestCreate_TeamMembership(t *testing.T) {
	membership := createNewMembership()
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
					testAccTeamExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "name", membership.TeamName),
				),
			},
		},
	})
}


func createNewMembership() MembershipData {
	replacementUser := os.Getenv("VO_REPLACEMENT_USERNAME")

	return MembershipData{
		User: victorops.User{
			FirstName: faker.FirstName(),
			LastName: faker.LastName(),
			Username: strings.ToLower(faker.Username()),
			Email: faker.Email(),
			Admin: true,
		},
		TeamName: faker.Word(),
		Replacement: replacementUser,
	}
}

func createMembershipResource(md MembershipData) string {
	return getTestTemplate("test_memb.tf", md)
}

func testMembershipDestroy(s *terraform.State) error {
	return nil
}