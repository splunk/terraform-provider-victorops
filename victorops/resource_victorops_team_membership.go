package victorops

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceTeamMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceTeamMembershipCreate,
		Read:   resourceTeamMembershipRead,
		Update: resourceTeamMembershipUpdate,
		Delete: resourceTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTeamMembershipImport,
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

func resourceTeamMembershipCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)

	// Make the request
	details, err := config.VictorOpsClient.AddTeamMember(teamid, username)
	if err != nil {
		return err
	}

	if details.StatusCode != 200 {
		return fmt.Errorf("failed to create user (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(teamid + "_" + username)
	return resourceTeamMembershipRead(d, m)
}

// Teams is a struct to parse the response of the team membership query into
type Teams struct {
	Teams []victorops.Team `json:"teams"`
}

func resourceTeamMembershipRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)

	isMember, details, err := config.VictorOpsClient.IsTeamMember(teamid, username)
	if err != nil {
		return err
	}

	if details.StatusCode == 404 || !isMember {
		d.SetId("")
		return nil
	} else if details.StatusCode != 200 {
		return fmt.Errorf("Failed to lookup team membership (%d): %s", details.StatusCode, details.ResponseBody)
	}

	return nil
}

func resourceTeamMembershipUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceTeamMembershipDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)
	teamid := d.Get("team_id").(string)
	replacement := d.Get("replacement_user").(string)

	if replacement == "" {
		return errors.New("replacement user must be specified to delete a team membership")
	}

	// Make the request
	details, err := config.VictorOpsClient.RemoveTeamMember(teamid, username, replacement)
	if err != nil {
		return err
	}

	if details.StatusCode != 200 {
		return fmt.Errorf("failed to remove %s from team %s (%d): %s", username, teamid, details.StatusCode, details.ResponseBody)
	}

	return nil
}

// resourceTeamMembershipImport imports an existing team membership.
// The import ID must be in the format "team_slug/username".
func resourceTeamMembershipImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid import ID %q: expected format \"team_slug/username\"", d.Id())
	}

	teamID := parts[0]
	username := parts[1]

	if err := d.Set("team_id", teamID); err != nil {
		return nil, err
	}
	if err := d.Set("user_name", username); err != nil {
		return nil, err
	}

	d.SetId(teamID + "_" + username)

	if err := resourceTeamMembershipRead(d, m); err != nil {
		return nil, err
	}

	if d.Id() == "" {
		return nil, fmt.Errorf("team membership %s/%s not found", teamID, username)
	}

	return []*schema.ResourceData{d}, nil
}
