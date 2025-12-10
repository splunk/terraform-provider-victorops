package victorops

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/victorops/go-victorops/victorops"
)

func resourceContact() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContactCreate,
		ReadContext:   resourceContactRead,
		DeleteContext: resourceContactDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceContactImport,
		},

		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the user this contact belongs to.",
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`(^([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})$)|(^[+\d- ]+$)`), "Value must be valid phone number or email address."),
				Description:  "The contact value (email address or phone number).",
			},
			"computed_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The contact value as formatted by the VictorOps API.",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A label for this contact (e.g., 'Work Phone', 'Personal Email').",
			},
			"internal_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The internal ID of the contact.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"phone", "email"}, false),
				Description:  "The type of contact: 'phone' or 'email'.",
			},
		},
	}
}

func typeToContactType(sType string) victorops.ContactType {
	if sType == "phone" {
		return victorops.GetContactTypes().Phone
	}
	return victorops.GetContactTypes().Email
}

func resourceContactCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	username := d.Get("user_name").(string)
	contactType := d.Get("type").(string)

	contact := &victorops.Contact{
		Label: d.Get("label").(string),
	}

	if contactType == "phone" {
		contact.PhoneNumber = d.Get("value").(string)
	} else if contactType == "email" {
		contact.Email = d.Get("value").(string)
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
		return diag.Errorf("failed to create contact (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("internal_id", newContact.ID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("computed_value", newContact.Value); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newContact.ExtID)

	return nil
}

func resourceContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Get("user_name").(string)
	contactID := d.Id()
	contactType := typeToContactType(d.Get("type").(string))

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	newContact, requestDetails, err := config.VictorOpsClient.GetContact(username, contactID, contactType)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode == 404 {
		d.SetId("")
		return diags
	}
	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to get contact (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("computed_value", newContact.Value); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("label", newContact.Label); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("internal_id", newContact.ID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceContactDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	username := d.Get("user_name").(string)
	contactID := d.Id()
	contactType := typeToContactType(d.Get("type").(string))

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	requestDetails, err := config.VictorOpsClient.DeleteContact(username, contactID, contactType)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to delete contact (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId("")
	return diags
}

func resourceContactImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 3)
	if len(idAttr) != 3 {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"username/contact_type/external_id\" for import", d.Id())
	}

	username := idAttr[0]
	contactType := idAttr[1]
	contactID := idAttr[2]

	d.Set("user_name", username)
	d.Set("type", contactType)
	d.SetId(contactID)

	diags := resourceContactRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to read contact during import")
	}

	if d.Id() == "" {
		return nil, fmt.Errorf("contact %s not found", idAttr)
	}

	d.Set("value", d.Get("computed_value").(string))

	return []*schema.ResourceData{d}, nil
}
