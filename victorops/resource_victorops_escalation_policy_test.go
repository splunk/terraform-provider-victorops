package victorops

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"os"
)

type PolicyData struct {
	TeamName     string
	PolicyName   string
	RotationSlug string
}

// Cannot be executed until the unmarshalling is fixed
// Error: json: cannot unmarshal array into Go value of type victorops.EscalationPolicy
//func TestEscalationPolicy_Create(t *testing.T) {
//	tfPolicyResourceName := "victorops_escalation_policy.test_policy"
//	pd := createNewPolicyModel()
//	resource.Test(t, resource.TestCase{
//		PreCheck: func() {
//			testAccPreCheck(t)
//		},
//		Providers:    testAccProviders,
//		CheckDestroy: testAccPolicyDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: createEscalationPolicyResource(pd),
//				Check: resource.ComposeTestCheckFunc(
//					testAccPolicyExists(tfPolicyResourceName),
//					resource.TestCheckResourceAttr(tfPolicyResourceName, "name", pd.PolicyName),
//				),
//			},
//		},
//	})
//}

func createNewPolicyModel() PolicyData {
	rotationSlug := os.Getenv("VO_ROTATION_GROUP_SLUG")
	return PolicyData{
		TeamName:     faker.Word(),
		PolicyName:   faker.Word(),
		RotationSlug: rotationSlug,
	}
}

func createEscalationPolicyResource(pd PolicyData) string {
	tmpl := getTestTemplate("test_policy.tf", pd)
	return tmpl
}

func testAccPolicyExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		policySlug := rs.Primary.ID
		apiClient := testAccProvider.Meta().(Config).VictorOpsClient
		_, _, err := apiClient.GetEscalationPolicy(policySlug)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccPolicyDestroy(s *terraform.State) error {
	return nil
}
