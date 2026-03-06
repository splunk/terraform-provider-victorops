package victorops

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// dataSourceEscalationPolicy returns a read-only data source for looking up an existing VictorOps escalation policy.
func dataSourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEscalationPolicyRead,

		Schema: map[string]*schema.Schema{
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug (unique identifier) of the escalation policy to look up.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name of the escalation policy.",
			},
			"team_slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The slug of the team this policy belongs to.",
			},
		},
	}
}

func dataSourceEscalationPolicyRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	slug := d.Get("slug").(string)

	policy, respDetails, err := config.VictorOpsClient.GetEscalationPolicy(slug)
	if err != nil {
		return err
	}

	if respDetails.StatusCode == 404 {
		return fmt.Errorf("escalation policy %q not found", slug)
	}

	if respDetails.StatusCode != 200 {
		return fmt.Errorf("error reading escalation policy %q (%d): %s", slug, respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(policy.ID)

	if err := d.Set("name", policy.Name); err != nil {
		return err
	}
	if err := d.Set("team_slug", policy.TeamID); err != nil {
		return err
	}

	return nil
}
