package victorops

import (
	"context"

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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The first name of the user.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The last name of the user.",
			},
			"user_name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The username of the user. This is immutable after creation.",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The email address of the user.",
			},
			"is_admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Deprecated:  "The admin parameter is deprecated in the VictorOps API. Use team admin assignments instead.",
				Description: "Whether the user is an admin. This field is deprecated and cannot be updated after creation.",
			},
			"expiration_hours": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     24,
				Description: "The number of hours until the user invitation expires. Only used during creation.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return true
				},
			},
			"replacement_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The username of a replacement user. Required when deleting a user.",
			},
			"default_email_contact_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the user's default email contact.",
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	username := d.Get("user_name").(string)

	user := &victorops.User{
		FirstName:       d.Get("first_name").(string),
		LastName:        d.Get("last_name").(string),
		Username:        username,
		Email:           d.Get("email").(string),
		Admin:           d.Get("is_admin").(bool),
		ExpirationHours: d.Get("expiration_hours").(int),
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	newUser, respDetails, err := config.VictorOpsClient.CreateUser(user)
	if err != nil {
		return diag.FromErr(err)
	}

	if respDetails.StatusCode != 200 {
		return diag.Errorf("failed to create user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId(newUser.Username)

	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Id()

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	user, respDetails, err := config.VictorOpsClient.GetUser(username)
	if err != nil {
		return diag.FromErr(err)
	}

	if respDetails.StatusCode == 404 {
		d.SetId("")
		return diags
	}

	if respDetails.StatusCode != 200 {
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
	if err := d.Set("user_name", user.Username); err != nil {
		return diag.FromErr(err)
	}

	// Wait for rate limiter before making second API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	defaultEmailContactID, requestDetails, err := config.VictorOpsClient.GetUserDefaultEmailContactID(username)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to get default email contact id for user (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("default_email_contact_id", defaultEmailContactID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	user := &victorops.User{
		FirstName: d.Get("first_name").(string),
		LastName:  d.Get("last_name").(string),
		Username:  d.Id(),
		Email:     d.Get("email").(string),
		Admin:     d.Get("is_admin").(bool),
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	user, respDetails, err := config.VictorOpsClient.UpdateUser(user)
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
		return diag.Errorf("replacement_user must be specified before a user can be deleted")
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	respDetails, err := config.VictorOpsClient.DeleteUser(d.Id(), replacementUser)
	if err != nil {
		return diag.FromErr(err)
	}

	if respDetails.StatusCode != 200 {
		return diag.Errorf("failed to delete user (%d): %s", respDetails.StatusCode, respDetails.ResponseBody)
	}

	d.SetId("")
	return diags
}
