package victorops

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The slug identifier for the team",
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
	}
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	// Create the team object for the request
	team := &victorops.Team{
		Name: d.Get("name").(string),
	}

	// Make the request
	newTeam, details, err := config.VictorOpsClient.CreateTeam(ctx, team)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to create team (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(newTeam.Slug)
	return resourceTeamRead(ctx, d, m)
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Make the request
	team, details, err := config.VictorOpsClient.GetTeam(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// If the team no longer exists then tell terraform that
	if details.StatusCode == 404 {
		d.SetId("")
		return diags
	} else if details.StatusCode != 200 {
		return diag.Errorf("failed to lookup team %s (%d): %s", d.Id(), details.StatusCode, details.ResponseBody)
	}

	if err := d.Set("name", team.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("slug", team.Slug); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("member_count", team.MemberCount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default_team", team.IsDefaultTeam); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	// Create the team object for the request
	team := &victorops.Team{
		Name: d.Get("name").(string),
	}

	// Make the request
	newTeam, details, err := config.VictorOpsClient.UpdateTeam(ctx, team)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to update team %s (%d): %s", d.Id(), details.StatusCode, details.ResponseBody)
	}

	if err := d.Set("name", newTeam.Name); err != nil {
		return diag.FromErr(err)
	}

	return resourceTeamRead(ctx, d, m)
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Make the request
	details, err := config.VictorOpsClient.DeleteTeam(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.FromErr(fmt.Errorf("failed to delete team %s (%d): %s", d.Id(), details.StatusCode, details.ResponseBody))
	}

	return diags
}
