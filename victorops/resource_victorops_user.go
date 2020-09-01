package victorops

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"first_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_admin": {
				Type:     schema.TypeBool,
				Required: true,
				// is_admin is not returned from the API on GET and can not be updated on PUT. So for now
				// we need to force recreation if this is changed (or it can be changed in the UI).
				ForceNew: true,
			},
			"expiration_hours": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  24,
				// Since this value can only ever be set at user creation and is never needed beyond that,
				// ignore all changes
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return true },
			},
			"replacement_user": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_email_contact_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)

	// Create the user object for the request
	user := &victorops.User{
		FirstName:       d.Get("first_name").(string),
		LastName:        d.Get("last_name").(string),
		Username:        username,
		Email:           d.Get("email").(string),
		Admin:           d.Get("is_admin").(bool),
		ExpirationHours: d.Get("expiration_hours").(int),
	}

	// Make the request
	newUser, respDetails, err := config.VictorOpsClient.CreateUser(user)
	if err != nil {
		return err
	}

	if respDetails.StatusCode != 200 {
		d.SetId("")
		return fmt.Errorf("failed to create user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(newUser.Username)

	return resourceUserRead(d, m)
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Id()

	// Make the request
	user, respDetails, err := config.VictorOpsClient.GetUser(username)
	if err != nil {
		return err
	}

	// If the user no longer exists then tell terraform that
	if respDetails.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	if respDetails.StatusCode != 200 {
		d.SetId("")
		return fmt.Errorf("error reading user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	err = d.Set("first_name", user.FirstName)
	if err != nil {
		return err
	}
	err = d.Set("last_name", user.LastName)
	if err != nil {
		return err
	}
	err = d.Set("email", user.Email)
	if err != nil {
		return err
	}

	// Also grab the default email contact id, for use in paging policies
	defaultEmailContactID, requestDetails, err := config.VictorOpsClient.GetUserDefaultEmailContactID(username)
	if err != nil {
		return err
	}

	if requestDetails.StatusCode != 200 {
		return fmt.Errorf("faile to get default email contact id for user (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	err = d.Set("default_email_contact_id", defaultEmailContactID)
	if err != nil {
		return err
	}

	err = d.Set("user_name", user.Username)
	if err != nil {
		return err
	}

	d.SetId(user.Username)
	err = d.Set("first_name", user.FirstName)
	if err != nil {
		return err
	}

	err = d.Set("last_name", user.LastName)
	if err != nil {
		return err
	}

	err = d.Set("email", user.Email)
	if err != nil {
		return err
	}

	// TODO: is_admin is not returned via the get API which means we cannot detect changes.
	// this is a bug that will affect imported users. Not sure how best to handle this yet

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Create the user object for the request
	user := &victorops.User{
		FirstName: d.Get("first_name").(string),
		LastName:  d.Get("last_name").(string),
		Username:  d.Id(),
		Email:     d.Get("email").(string),
		Admin:     d.Get("is_admin").(bool),
	}

	// Make the request
	user, respDetails, err := config.VictorOpsClient.UpdateUser(user)
	if err != nil {
		return err
	}

	if respDetails.StatusCode != 200 {
		return fmt.Errorf("failed to update user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(user.Username)
	return resourceUserRead(d, m)
}

func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	replacementUser := d.Get("replacement_user").(string)

	if replacementUser == "" {
		return errors.New("replacement_user must be specified before a user can be deleted")
	}

	// Make the request
	respDetails, err := config.VictorOpsClient.DeleteUser(d.Id(), replacementUser)
	if err != nil {
		return err
	}

	if respDetails.StatusCode != 200 {
		d.SetId("")
		return fmt.Errorf("failed to delete user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	return nil
}
