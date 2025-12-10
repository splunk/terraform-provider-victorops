package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamAdmins() *schema.Resource {
	return &schema.Resource{
		Description: "Get the admins for a team.",
		ReadContext: dataSourceTeamAdminsRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug/ID of the team.",
			},
			"admins": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The team admins.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the admin.",
						},
					},
				},
			},
		},
	}
}

func dataSourceTeamAdminsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	teamID := d.Get("team_id").(string)

	admins, err := apiClient.GetTeamAdmins(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(teamID)

	adminList := make([]map[string]interface{}, len(admins))
	for i, a := range admins {
		adminList[i] = map[string]interface{}{
			"username": a.Username,
		}
	}

	if err := d.Set("admins", adminList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
