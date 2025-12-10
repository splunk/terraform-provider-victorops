package victorops

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceTeamMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamMembershipCreate,
		ReadContext:   resourceTeamMembershipRead,
		UpdateContext: resourceTeamMembershipUpdate,
		DeleteContext: resourceTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTeamMembershipImport,
		},

		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the user to add to the team.",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The slug/ID of the team.",
			},
			"replacement_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The username of a replacement user. Required when deleting a membership.",
			},
		},
	}
}

// Teams is a struct to parse the response of the team membership query
type Teams struct {
	Teams []victorops.Team `json:"teams"`
}

func resourceTeamMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	details, err := config.VictorOpsClient.AddTeamMember(teamid, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to add user %s to team %s (%d): %s", username, teamid, details.StatusCode, details.ResponseBody)
	}

	d.SetId(teamid + "_" + username)
	return resourceTeamMembershipRead(ctx, d, m)
}

func resourceTeamMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	isMember, details, err := config.VictorOpsClient.IsTeamMember(teamid, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode == 404 || !isMember {
		d.SetId("")
		return diags
	}
	if details.StatusCode != 200 {
		return diag.Errorf("failed to lookup team membership (%d): %s", details.StatusCode, details.ResponseBody)
	}

	return diags
}

func resourceTeamMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceTeamMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)
	replacement := d.Get("replacement_user").(string)

	if replacement == "" {
		return diag.Errorf("replacement_user must be specified to delete a team membership")
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	details, err := config.VictorOpsClient.RemoveTeamMember(teamid, username, replacement)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to remove %s from team %s (%d): %s", username, teamid, details.StatusCode, details.ResponseBody)
	}

	d.SetId("")
	return diags
}

func resourceTeamMembershipImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)
	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"team_id/username\" for import", d.Id())
	}

	teamID := idAttr[0]
	username := idAttr[1]

	d.Set("team_id", teamID)
	d.Set("user_name", username)
	d.SetId(teamID + "_" + username)

	return []*schema.ResourceData{d}, nil
}
