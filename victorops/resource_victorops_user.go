package victorops

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	newUser, respDetails, err := config.VictorOpsClient.CreateUser(ctx, user)
	if err != nil {
		return diag.FromErr(err)
	}

	if respDetails.StatusCode != 200 {
		d.SetId("")
		return diag.Errorf("failed to create user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(newUser.Username)

	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Id()

	// Make the request
	user, respDetails, err := config.VictorOpsClient.GetUser(ctx, username)
	if err != nil {
		return diag.FromErr(err)
	}

	// If the user no longer exists then tell terraform that
	if respDetails.StatusCode == 404 {
		d.SetId("")
		return diags
	}

	if respDetails.StatusCode != 200 {
		d.SetId("")
		return diag.Errorf("error reading user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	if err := d.Set("first_name", user.FirstName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_name", user.LastName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}

	// Also grab the default email contact id, for use in paging policies
	defaultEmailContactID, requestDetails, err := config.VictorOpsClient.GetUserDefaultEmailContactID(ctx, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to get default email contact id for user (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("default_email_contact_id", defaultEmailContactID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("user_name", user.Username); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Username)

	// TODO: is_admin is not returned via the get API which means we cannot detect changes.
	// this is a bug that will affect imported users. Not sure how best to handle this yet

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	user, respDetails, err := config.VictorOpsClient.UpdateUser(ctx, user)
	if err != nil {
		return diag.FromErr(err)
	}

	if respDetails.StatusCode != 200 {
		return diag.Errorf("failed to update user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(user.Username)
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	replacementUser := d.Get("replacement_user").(string)

	if replacementUser == "" {
		return diag.FromErr(errors.New("replacement_user must be specified before a user can be deleted"))
	}

	// Make the request
	respDetails, err := config.VictorOpsClient.DeleteUser(ctx, d.Id(), replacementUser)
	if err != nil {
		return diag.FromErr(err)
	}

	if respDetails.StatusCode != 200 {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("failed to delete user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody))
	}

	return diags
}
