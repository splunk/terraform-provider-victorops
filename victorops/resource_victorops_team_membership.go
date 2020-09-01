package victorops

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceTeamMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceTeamMembershipCreate,
		Read:   resourceTeamMembershipRead,
		Update: resourceTeamMembershipUpdate,
		Delete: resourceTeamMembershipDelete,

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
