package victorops

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceUserRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	// Re-use VO_REPLACEMENT_USERNAME as a known-existing user for the data source test.
	username := os.Getenv("VO_REPLACEMENT_USERNAME")
	if username == "" {
		t.Skip("VO_REPLACEMENT_USERNAME must be set to run data source user acceptance test")
	}

	tfDataSourceName := "data.victorops_user.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserConfig(username),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUserExists(tfDataSourceName),
					resource.TestCheckResourceAttr(tfDataSourceName, "user_name", username),
					resource.TestCheckResourceAttrSet(tfDataSourceName, "first_name"),
					resource.TestCheckResourceAttrSet(tfDataSourceName, "last_name"),
					resource.TestCheckResourceAttrSet(tfDataSourceName, "email"),
				),
			},
		},
	})
}

func testAccDataSourceUserConfig(username string) string {
	return fmt.Sprintf(`
data "victorops_user" "test" {
  user_name = "%s"
}
`, username)
}

func testAccDataSourceUserExists(dataSourceName string) resource.TestCheckFunc {
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
