package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRead,

		Schema: map[string]*schema.Schema{
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The team slug to look up",
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
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The version of the team",
			},
			"is_default_team": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is the default team",
			},
		},
	}
}

func dataSourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	slug := d.Get("slug").(string)

	team, details, err := config.VictorOpsClient.GetTeam(ctx, slug)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode == 404 {
		return diag.Errorf("team %s not found", slug)
	} else if details.StatusCode != 200 {
		return diag.Errorf("failed to get team (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(team.Slug)
	if err := d.Set("name", team.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("member_count", team.MemberCount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", team.Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default_team", team.IsDefaultTeam); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
