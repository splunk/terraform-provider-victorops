package victorops

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/victorops/go-victorops/victorops"
	"os"
	"regexp"
	"strings"
	"testing"
)

type UserData struct {
	User victorops.User
	Replacement string
}


func TestUser_Create(t *testing.T) {
	replacementUsername := os.Getenv("VO_REPLACEMENT_USERNAME")
	voUsr := createNewUser()
	tfResourceName := "victorops_user.test_user"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: createUserResource(voUsr, replacementUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "user_name", voUsr.Username),
					resource.TestCheckResourceAttr(tfResourceName, "first_name", voUsr.FirstName),
					resource.TestCheckResourceAttr(tfResourceName, "last_name", voUsr.LastName),
					resource.TestCheckResourceAttr(tfResourceName, "email", voUsr.Email),
					resource.TestCheckResourceAttr(tfResourceName, "is_admin", "true"),
					resource.TestCheckResourceAttr(tfResourceName, "replacement_user", replacementUsername),
				),
			},
		},
	})
}


func TestUser_Update(t *testing.T) {
	replacementUsername := os.Getenv("VO_REPLACEMENT_USERNAME")
	createUser := createNewUser()
	updateUser := updateUser(createUser)
	tfResourceName := "victorops_user.test_user"
	resource.Test(t, resource.TestCase {
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: createUserResource(createUser, replacementUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "user_name", createUser.Username),
					resource.TestCheckResourceAttr(tfResourceName, "first_name", createUser.FirstName),
					resource.TestCheckResourceAttr(tfResourceName, "last_name", createUser.LastName),
					resource.TestCheckResourceAttr(tfResourceName, "email", createUser.Email),
					resource.TestCheckResourceAttr(tfResourceName, "is_admin", "true"),
					resource.TestCheckResourceAttr(tfResourceName, "replacement_user", replacementUsername),
				),
			},
			{
				Config: createUserResource(updateUser, replacementUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(tfResourceName),
					resource.TestCheckResourceAttr(tfResourceName, "user_name", updateUser.Username),
					resource.TestCheckResourceAttr(tfResourceName, "first_name", updateUser.FirstName),
					resource.TestCheckResourceAttr(tfResourceName, "last_name", updateUser.LastName),
					resource.TestCheckResourceAttr(tfResourceName, "email", updateUser.Email),
					resource.TestCheckResourceAttr(tfResourceName, "is_admin", "true"),
					resource.TestCheckResourceAttr(tfResourceName, "replacement_user", replacementUsername),
				),
			},
		},
	})
}

func createNewUser() victorops.User {
	return victorops.User{
		FirstName: faker.FirstName(),
		LastName: faker.LastName(),
		Username: strings.ToLower(faker.Username()),
		Email: faker.Email(),
		Admin: true,
	}
}

func updateUser(user victorops.User) victorops.User {
	user.FirstName = faker.FirstName()
	user.LastName = faker.LastName()
	return user
}

func createUserResource(user victorops.User, replName string) string {
	ud := UserData{
		User: user,
		Replacement: replName,
	}
	return getTestTemplate("test_user.tf", ud)
}

func testAccUserExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		username := rs.Primary.ID
		apiClient := testAccProvider.Meta().(Config).VictorOpsClient
		_, _, err := apiClient.GetUser(username)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccUserDestroy(s *terraform.State) error {

	apiClient := testAccProvider.Meta().(Config).VictorOpsClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "victorops_user" {
			continue
		}

		_, response, _ := apiClient.GetUser(rs.Primary.ID)
		notFoundErr := "None found for username:*"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(response.ResponseBody)) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, response.ResponseBody)
		}
	}

	return nil
}
