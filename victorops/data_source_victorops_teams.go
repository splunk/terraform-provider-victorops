package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamsRead,

		Schema: map[string]*schema.Schema{
			"teams": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all teams",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"slug": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The team slug",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The team's name",
						},
						"member_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of members in the team",
						},
						"is_default_team": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this is the default team",
						},
					},
				},
			},
		},
	}
}

func dataSourceTeamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	teamsResp, details, err := config.VictorOpsClient.GetAllTeams(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get teams (%d): %s", details.StatusCode, details.ResponseBody)
	}

	// teamsResp is *[]Team, so dereference it
	teamsList := *teamsResp
	teams := make([]map[string]interface{}, len(teamsList))
	for i, team := range teamsList {
		teams[i] = map[string]interface{}{
			"slug":            team.Slug,
			"name":            team.Name,
			"member_count":    team.MemberCount,
			"is_default_team": team.IsDefaultTeam,
		}
	}

	d.SetId("victorops_teams")
	if err := d.Set("teams", teams); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
