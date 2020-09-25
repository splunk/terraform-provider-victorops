package victorops

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"os"
)

type RoutingKeyData struct {
	TeamName       string
	PolicyName     string
	RotationSlug   string
	RoutingKeyName string
}

// Cannot be tested until api can remove routing keys
// Error: errors during apply: deleting routing keys not yet implemented in the API, please delete in the UI
//func TestRoutingKey_Create(t *testing.T) {
//	tfResourceName := "victorops_routing_key.test_key"
//	rkd := createRouteKeyModel()
//	resource.Test(t, resource.TestCase{
//		PreCheck: func() {
//			testAccPreCheck(t)
//		},
//		Providers:    testAccProviders,
//		PreventPostDestroyRefresh: true,
//		Steps: []resource.TestStep{
//			{
//				Config: createRoutKeyResource(rkd),
//				Check: testAccRouteKeyExists(tfResourceName),
//			},
//		},
//	})
//}


func createRouteKeyModel() RoutingKeyData {
	rs := os.Getenv("VO_ROTATION_GROUP_SLUG")
	return RoutingKeyData{
		TeamName:       faker.Word(),
		PolicyName:     faker.Word(),
		RotationSlug:   rs,
		RoutingKeyName: faker.Word(),
	}
}

func createRoutKeyResource(rk RoutingKeyData) string {
	return getTestTemplate("test_routing_key.tf", rk)
}


func testAccRouteKeyExists(resource string) resource.TestCheckFunc {
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
		_, _, err := apiClient.GetRoutingKey(username)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccRouteKeyDestroy(s *terraform.State) error {

	//apiClient := testAccProvider.Meta().(Config).VictorOpsClient
	//
	//for _, rs := range s.RootModule().Resources {
	//	if rs.Type != "victorops_routing_key" {
	//		continue
	//	}
	//
	//	_, response, _ := apiClient.GetUser(rs.Primary.ID)
	//	notFoundErr := "None found for username:*"
	//	expectedErr := regexp.MustCompile(notFoundErr)
	//	if !expectedErr.Match([]byte(response.ResponseBody)) {
	//		return fmt.Errorf("expected %s, got %s", notFoundErr, response.ResponseBody)
	//	}
	//}

	return nil
}