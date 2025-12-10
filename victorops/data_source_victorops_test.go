package victorops

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUsersConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.victorops_users.all", "users.#"),
				),
			},
		},
	})
}

func testAccDataSourceUsersConfig() string {
	return `
data "victorops_users" "all" {}
`
}

func TestAccDataSourceRoutingKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRoutingKeysConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.victorops_routing_keys.all", "routing_keys.#"),
				),
			},
		},
	})
}

func testAccDataSourceRoutingKeysConfig() string {
	return `
data "victorops_routing_keys" "all" {}
`
}

func TestAccDataSourceTeamAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTeamAdminsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.victorops_team_admins.test", "admins.#"),
				),
			},
		},
	})
}

func testAccDataSourceTeamAdminsConfig() string {
	return `
resource "victorops_team" "test_for_admins" {
  name = "test-admins-team"
}

data "victorops_team_admins" "test" {
  team_id = victorops_team.test_for_admins.id
}
`
}

func TestAccDataSourceUserDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	username := os.Getenv("VO_REPLACEMENT_USERNAME")
	if username == "" {
		t.Skip("VO_REPLACEMENT_USERNAME not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserDevicesConfig(username),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.victorops_user_devices.test", "devices.#"),
				),
			},
		},
	})
}

func testAccDataSourceUserDevicesConfig(username string) string {
	return fmt.Sprintf(`
data "victorops_user_devices" "test" {
  username = "%s"
}
`, username)
}

func TestAccDataSourceRotations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRotationsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.victorops_rotations.test", "rotations.#"),
				),
			},
		},
	})
}

func testAccDataSourceRotationsConfig() string {
	return `
resource "victorops_team" "test_for_rotations" {
  name = "test-rotations-team"
}

data "victorops_rotations" "test" {
  team_id = victorops_team.test_for_rotations.id
}
`
}

func TestAccDataSourceTeamOncallSchedule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTeamOncallScheduleConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.victorops_team_oncall_schedule.test", "schedules.#"),
				),
			},
		},
	})
}

func testAccDataSourceTeamOncallScheduleConfig() string {
	return `
resource "victorops_team" "test_for_oncall" {
  name = "test-oncall-team"
}

data "victorops_team_oncall_schedule" "test" {
  team_id      = victorops_team.test_for_oncall.id
  days_forward = 7
}
`
}
