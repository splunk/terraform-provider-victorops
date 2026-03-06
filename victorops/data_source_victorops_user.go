package victorops

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// dataSourceUser returns a read-only data source for looking up an existing VictorOps user.
func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username of the VictorOps user to look up.",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's first name.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's last name.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's email address.",
			},
			"default_email_contact_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The internal ID of the user's default email contact method.",
			},
		},
	}
}

func dataSourceUserRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)

	user, respDetails, err := config.VictorOpsClient.GetUser(username)
	if err != nil {
		return err
	}

	if respDetails.StatusCode == 404 {
		return fmt.Errorf("user %q not found", username)
	}

	if respDetails.StatusCode != 200 {
		return fmt.Errorf("error reading user %q (%d): %s", username, respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(user.Username)

	if err := d.Set("first_name", user.FirstName); err != nil {
		return err
	}
	if err := d.Set("last_name", user.LastName); err != nil {
		return err
	}
	if err := d.Set("email", user.Email); err != nil {
		return err
	}

	defaultEmailContactID, reqDetails, err := config.VictorOpsClient.GetUserDefaultEmailContactID(username)
	if err != nil {
		return err
	}

	if reqDetails.StatusCode != 200 {
		return fmt.Errorf("error reading default email contact id for user %q (%d): %s", username, reqDetails.StatusCode, reqDetails.ResponseBody)
	}

	if err := d.Set("default_email_contact_id", defaultEmailContactID); err != nil {
		return err
	}

	return nil
}
