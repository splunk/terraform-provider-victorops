package victorops

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceTeamRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	// A team is created in the same test config so we have a known slug to look up.
	teamName := "TF-DataSource-Team"
	tfResourceName := "victorops_team.ds_test_team"
	tfDataSourceName := "data.victorops_team.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTeamConfig(teamName),
				Check: resource.ComposeTestCheckFunc(
					testAccTeamExists(tfResourceName),
					testAccDataSourceTeamExists(tfDataSourceName),
					resource.TestCheckResourceAttr(tfDataSourceName, "name", teamName),
					resource.TestCheckResourceAttrSet(tfDataSourceName, "slug"),
				),
			},
		},
	})
}

func testAccDataSourceTeamConfig(teamName string) string {
	return fmt.Sprintf(`
resource "victorops_team" "ds_test_team" {
  name = "%s"
}

data "victorops_team" "test" {
  slug = victorops_team.ds_test_team.id
}
`, teamName)
}

func testAccDataSourceTeamExists(dataSourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for data source %s", dataSourceName)
		}
		return nil
	}
}
