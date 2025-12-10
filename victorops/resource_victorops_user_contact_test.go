package victorops

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUserContactEmailCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	username := os.Getenv("VO_REPLACEMENT_USERNAME")
	if username == "" {
		t.Skip("VO_REPLACEMENT_USERNAME not set")
	}

	tfResourceName := "victorops_user_contact_email.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		CheckDestroy: testAccUserContactEmailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserContactEmailConfig(username),
				Check: resource.ComposeTestCheckFunc(
					testAccUserContactEmailExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "username", username),
					resource.TestCheckResourceAttr(tfResourceName, "label", "Test Email"),
					resource.TestCheckResourceAttrSet(tfResourceName, "ext_id"),
				),
			},
		},
	})
}

func testAccUserContactEmailConfig(username string) string {
	return fmt.Sprintf(`
resource "victorops_user_contact_email" "test" {
  username = "%s"
  label    = "Test Email"
  email    = "tf-test-%s@example.com"
}
`, username, username)
}

func testAccUserContactEmailExists(resourceName string) resource.TestCheckFunc {
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

func testAccUserContactEmailDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_user_contact_email" {
			continue
		}
	}
	return nil
}

func TestAccUserContactPhoneCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	username := os.Getenv("VO_REPLACEMENT_USERNAME")
	if username == "" {
		t.Skip("VO_REPLACEMENT_USERNAME not set")
	}

	tfResourceName := "victorops_user_contact_phone.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"victorops": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		CheckDestroy: testAccUserContactPhoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserContactPhoneConfig(username),
				Check: resource.ComposeTestCheckFunc(
					testAccUserContactPhoneExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "username", username),
					resource.TestCheckResourceAttr(tfResourceName, "label", "Test Phone"),
					resource.TestCheckResourceAttrSet(tfResourceName, "ext_id"),
				),
			},
		},
	})
}

func testAccUserContactPhoneConfig(username string) string {
	return fmt.Sprintf(`
resource "victorops_user_contact_phone" "test" {
  username = "%s"
  label    = "Test Phone"
  phone    = "+1-555-123-4567"
}
`, username)
}

func testAccUserContactPhoneExists(resourceName string) resource.TestCheckFunc {
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

func testAccUserContactPhoneDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_user_contact_phone" {
			continue
		}
	}
	return nil
}
