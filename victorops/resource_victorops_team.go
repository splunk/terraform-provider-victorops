package victorops

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceTeamCreate,
		Read:   resourceTeamRead,
		Update: resourceTeamUpdate,
		Delete: resourceTeamDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTeamCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Create the user object for the request
	team := &victorops.Team{
		Name: d.Get("name").(string),
	}

	// Make the request
	newTeam, details, err := config.VictorOpsClient.CreateTeam(team)
	if err != nil {
		return err
	}

	if details.StatusCode != 200 {
		return fmt.Errorf("failed to create user (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(newTeam.Slug)
	return resourceTeamRead(d, m)
}

func resourceTeamRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Make the request
	team, details, err := config.VictorOpsClient.GetTeam(d.Id())
	if err != nil {
		return err
	}

	// If the phone no longer exists then tell terraform that
	if details.StatusCode == 404 {
		d.SetId("")
		return nil
	} else if details.StatusCode != 200 {
		return fmt.Errorf("failed to lookup team %s (%d): %s", d.Id(), details.StatusCode, details.ResponseBody)
	}

	err = d.Set("name", team.Name)
	if err != nil {
		return err
	}

	return nil
}

func resourceTeamUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Create the user object for the request
	team := &victorops.Team{
		Name: d.Get("name").(string),
	}

	// Make the request
	newTeam, details, err := config.VictorOpsClient.UpdateTeam(team)
	if err != nil {
		return err
	}

	if details.StatusCode != 200 {
		return fmt.Errorf("failed to update team %s (%d): %s", d.Id(), details.StatusCode, details.ResponseBody)
	}

	err = d.Set("name", newTeam.Name)
	if err != nil {
		return err
	}

	return resourceTeamRead(d, m)
}

func resourceTeamDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Make the request
	details, err := config.VictorOpsClient.DeleteTeam(d.Id())
	if err != nil {
		return err
	}

	if details.StatusCode != 200 {
		return fmt.Errorf("failed to delete team %s (%d): %s", d.Id(), details.StatusCode, details.ResponseBody)
	}

	return nil
}
