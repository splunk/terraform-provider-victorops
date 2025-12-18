package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEscalationPolicyRead,

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The escalation policy ID to look up",
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
			"ignore_custom_paging_policies": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether to ignore custom paging policies",
			},
		},
	}
}

func dataSourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	policyID := d.Get("policy_id").(string)

	policy, details, err := config.VictorOpsClient.GetEscalationPolicy(ctx, policyID)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode == 404 {
		return diag.Errorf("escalation policy %s not found", policyID)
	} else if details.StatusCode != 200 {
		return diag.Errorf("failed to get escalation policy (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(policy.ID)
	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team_id", policy.TeamID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ignore_custom_paging_policies", policy.IgnoreCustomPagingPolicies); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
