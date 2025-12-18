package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username to look up",
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
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "When the user was created",
			},
			"verified": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user is verified",
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	username := d.Get("user_name").(string)

	user, details, err := config.VictorOpsClient.GetUser(ctx, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode == 404 {
		return diag.Errorf("user %s not found", username)
	} else if details.StatusCode != 200 {
		return diag.Errorf("failed to get user (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(user.Username)
	if err := d.Set("first_name", user.FirstName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_name", user.LastName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", user.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("verified", user.Verified); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
