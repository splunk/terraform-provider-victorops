package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEscalationPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEscalationPoliciesRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional team ID to filter policies by",
			},
			"policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of escalation policies",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy name",
						},
						"team_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The team ID this policy belongs to",
						},
						"team_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The team name this policy belongs to",
						},
					},
				},
			},
		},
	}
}

func dataSourceEscalationPoliciesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	filterTeamID := d.Get("team_id").(string)

	policiesResp, details, err := config.VictorOpsClient.GetAllEscalationPolicies(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get escalation policies (%d): %s", details.StatusCode, details.ResponseBody)
	}

	policies := make([]map[string]interface{}, 0)
	for _, policyElement := range policiesResp.Policies {
		// Filter by team_id if specified
		if filterTeamID != "" && policyElement.Team.Slug != filterTeamID {
			continue
		}

		policies = append(policies, map[string]interface{}{
			"id":        policyElement.Policy.Slug,
			"name":      policyElement.Policy.Name,
			"team_id":   policyElement.Team.Slug,
			"team_name": policyElement.Team.Name,
		})
	}

	d.SetId("victorops_escalation_policies")
	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
