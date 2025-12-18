package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,

		Schema: map[string]*schema.Schema{
			"users": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all users",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's first name",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's last name",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's email address",
						},
						"verified": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user is verified",
						},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Use v2 API which returns a flat list
	usersResp, details, err := config.VictorOpsClient.GetAllUserV2(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get users (%d): %s", details.StatusCode, details.ResponseBody)
	}

	users := make([]map[string]interface{}, len(usersResp.Users))
	for i, user := range usersResp.Users {
		users[i] = map[string]interface{}{
			"user_name":  user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"verified":   user.Verified,
		}
	}

	d.SetId("victorops_users")
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
