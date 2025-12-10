package victorops

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/victorops/go-victorops/victorops"
)

func resourceUserContactEmail() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages an email contact method for a VictorOps user.",
		CreateContext: resourceUserContactEmailCreate,
		ReadContext:   resourceUserContactEmailRead,
		DeleteContext: resourceUserContactEmailDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserContactEmailImport,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the user this contact email belongs to.",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A label for this email contact (e.g., 'Work Email', 'Personal Email').",
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The email address.",
			},
			"rank": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     0,
				Description: "The rank (priority) of this contact method.",
			},
			"ext_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external ID of the contact.",
			},
		},
	}
}

func resourceUserContactEmailCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	username := d.Get("username").(string)

	contact := &victorops.Contact{
		Label: d.Get("label").(string),
		Email: d.Get("email").(string),
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	newContact, requestDetails, err := config.VictorOpsClient.CreateContact(username, contact)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to create contact email (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("ext_id", newContact.ExtID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", username, newContact.ExtID))

	return resourceUserContactEmailRead(ctx, d, m)
}

func resourceUserContactEmailRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	username := d.Get("username").(string)
	extID := d.Get("ext_id").(string)
	if extID == "" {
		parts := strings.SplitN(d.Id(), "/", 2)
		if len(parts) == 2 {
			username = parts[0]
			extID = parts[1]
		}
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	contact, requestDetails, err := config.VictorOpsClient.GetContact(username, extID, victorops.GetContactTypes().Email)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode == 404 {
		d.SetId("")
		return diags
	}
	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to get contact email (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("label", contact.Label); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ext_id", contact.ExtID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", username); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserContactEmailDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	username := d.Get("username").(string)
	extID := d.Get("ext_id").(string)

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	requestDetails, err := config.VictorOpsClient.DeleteContact(username, extID, victorops.GetContactTypes().Email)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to delete contact email (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId("")
	return diags
}

func resourceUserContactEmailImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)
	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"username/ext_id\" for import", d.Id())
	}

	username := idAttr[0]
	extID := idAttr[1]

	d.Set("username", username)
	d.Set("ext_id", extID)
	d.SetId(fmt.Sprintf("%s/%s", username, extID))

	return []*schema.ResourceData{d}, nil
}
