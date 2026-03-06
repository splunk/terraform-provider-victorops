package victorops

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// dataSourceTeam returns a read-only data source for looking up an existing VictorOps team.
func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTeamRead,

		Schema: map[string]*schema.Schema{
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug (unique identifier) of the VictorOps team to look up.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name of the team.",
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of members on the team.",
			},
			"is_default_team": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is the default team for the organization.",
			},
		},
	}
}

func dataSourceTeamRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	slug := d.Get("slug").(string)

	team, respDetails, err := config.VictorOpsClient.GetTeam(slug)
	if err != nil {
		return err
	}

	if respDetails.StatusCode == 404 {
		return fmt.Errorf("team %q not found", slug)
	}

	if respDetails.StatusCode != 200 {
		return fmt.Errorf("error reading team %q (%d): %s", slug, respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(team.Slug)

	if err := d.Set("name", team.Name); err != nil {
		return err
	}
	if err := d.Set("member_count", team.MemberCount); err != nil {
		return err
	}
	if err := d.Set("is_default_team", team.IsDefaultTeam); err != nil {
		return err
	}

	return nil
}
