package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Get users for the organization.",
		ReadContext: dataSourceUsersRead,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional email address to filter users (must be at least 3 characters).",
			},
			"users": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The users.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username.",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's first name.",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's last name.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's email address.",
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
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	email := d.Get("email").(string)

	users, err := apiClient.GetUsers(email)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("users")

	userList := make([]map[string]interface{}, len(users))
	for i, u := range users {
		userList[i] = map[string]interface{}{
			"username":   u.Username,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"email":      u.Email,
		}
	}

	if err := d.Set("users", userList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
