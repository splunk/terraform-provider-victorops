package victorops

import (
	"context"
	"errors"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"replacement_user": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceTeamMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)

	// Make the request
	details, err := config.VictorOpsClient.AddTeamMember(ctx, teamid, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to create team membership (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(teamid + "/" + username)
	return resourceTeamMembershipRead(ctx, d, m)
}

// Teams is a struct to parse the response of the team membership query into
type Teams struct {
	Teams []victorops.Team `json:"teams"`
}

func resourceTeamMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)

	isMember, details, err := config.VictorOpsClient.IsTeamMember(ctx, teamid, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode == 404 || !isMember {
		d.SetId("")
		return diags
	} else if details.StatusCode != 200 {
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
		return diag.FromErr(errors.New("replacement user must be specified to delete a team membership"))
	}

	// Make the request
	details, err := config.VictorOpsClient.RemoveTeamMember(ctx, teamid, username, replacement)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.FromErr(fmt.Errorf("failed to remove %s from team %s (%d): %s", username, teamid, details.StatusCode, details.ResponseBody))
	}

	return diags
}

// resourceTeamMembershipImport imports a team membership using the format "{team_slug}/{username}"
func resourceTeamMembershipImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Parse "{team_slug}/{username}"
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"team_slug/username\" for import", d.Id())
	}

	teamID := parts[0]
	username := parts[1]

	if err := d.Set("team_id", teamID); err != nil {
		return nil, err
	}
	if err := d.Set("user_name", username); err != nil {
		return nil, err
	}

	// Validate the membership exists
	config := m.(Config)
	isMember, details, err := config.VictorOpsClient.IsTeamMember(ctx, teamID, username)
	if err != nil {
		return nil, err
	}

	if details.StatusCode == 404 || !isMember {
		return nil, fmt.Errorf("user %s is not a member of team %s", username, teamID)
	}

	d.SetId(teamID + "/" + username)
	return []*schema.ResourceData{d}, nil
}
