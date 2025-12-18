package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOnCall() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOnCallRead,

		Schema: map[string]*schema.Schema{
			"teams_on_call": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of teams and their on-call information",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"team_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The team name",
						},
						"team_slug": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The team slug",
						},
						"on_call_users": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Users currently on-call for this team",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The username of the on-call user",
									},
									"escalation_policy_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The escalation policy name",
									},
									"escalation_policy_slug": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The escalation policy slug",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceOnCallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	onCallResp, details, err := config.VictorOpsClient.GetOnCallCurrent(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get on-call information (%d): %s", details.StatusCode, details.ResponseBody)
	}

	teamsOnCall := make([]map[string]interface{}, len(onCallResp.TeamsOnCall))
	for i, team := range onCallResp.TeamsOnCall {
		onCallUsers := make([]map[string]interface{}, 0)
		for _, policy := range team.OnCall {
			for _, user := range policy.Users {
				onCallUsers = append(onCallUsers, map[string]interface{}{
					"username":               user.OnCallUser.Username,
					"escalation_policy_name": policy.EscalationPolicy.Name,
					"escalation_policy_slug": policy.EscalationPolicy.Slug,
				})
			}
		}

		teamsOnCall[i] = map[string]interface{}{
			"team_name":     team.Team.Name,
			"team_slug":     team.Team.Slug,
			"on_call_users": onCallUsers,
		}
	}

	d.SetId("victorops_on_call_current")
	if err := d.Set("teams_on_call", teamsOnCall); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
