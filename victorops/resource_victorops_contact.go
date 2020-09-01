package victorops

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/victorops/go-victorops/victorops"
)

func resourceContact() *schema.Resource {
	return &schema.Resource{
		Create: resourceContactCreate,
		Read:   resourceContactRead,
		Delete: resourceContactDelete,
		Importer: &schema.ResourceImporter{
			State: resourceContactImport,
		},

		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`(^([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})$)|(^[+\d- ]+$)`), "Value must be valid phone number or email address."),
			},
			"computed_value": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"label": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"internal_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"phone", "email"}, false),
			},
		},
	}
}

func typeToContactType(sType string) victorops.ContactType {
	if sType == "phone" {
		return victorops.GetContactTypes().Phone
	} else {
		return victorops.GetContactTypes().Email
	}
}

func resourceContactCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)
	contactType := d.Get("type").(string)

	// Create the object for the request
	contact := &victorops.Contact{
		Label: d.Get("label").(string),
	}

	if contactType == "phone" {
		contact.PhoneNumber = d.Get("value").(string)
	} else if contactType == "email" {
		contact.Email = d.Get("value").(string)
	}

	// Make the request
	newContact, requestDetails, err := config.VictorOpsClient.CreateContact(username, contact)
	if err != nil {
		return err
	}

	if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to create contact (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	err = d.Set("internal_id", newContact.ID)
	if err != nil {
		return err
	}

	err = d.Set("computed_value", newContact.Value)
	if err != nil {
		return err
	}

	d.SetId(newContact.ExtID)

	return nil
}

func resourceContactRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)
	contactID := d.Id()
	contactType := typeToContactType(d.Get("type").(string))

	// Make the request
	newContact, requestDetails, err := config.VictorOpsClient.GetContact(username, contactID, contactType)
	if err != nil {
		return err
	}

	// If the contact no longer exists then tell terraform that
	if requestDetails.StatusCode == 404 {
		d.SetId("")
		return nil
	} else if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to get contact (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	// We compare the computed value against the saved state instead of the user supplied value
	// as the cpmputed value is the value as formatted by the victorops API on creation.
	err = d.Set("computed_value", newContact.Value)
	if err != nil {
		return err
	}

	err = d.Set("label", newContact.Label)
	if err != nil {
		return err
	}

	err = d.Set("internal_id", newContact.ID)
	if err != nil {
		return err
	}

	return nil
}

func resourceContactDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("user_name").(string)
	contactID := d.Id()
	contactType := typeToContactType(d.Get("type").(string))

	// Make the request
	requestDetails, err := config.VictorOpsClient.DeleteContact(username, contactID, contactType)
	if err != nil {
		return err
	}

	if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to get delete contact (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	return nil
}

// Because we are unable to read a contact from the victorops public API given only the ID of that contact,
// this method takes in a / delimited string that contains all of the information we need to read in a contact
// and triggers the read method
func resourceContactImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 3)
	var username string
	var contactType string
	var contactID string
	if len(idAttr) == 3 {
		username = idAttr[0]
		contactType = idAttr[1]
		contactID = idAttr[2]
	} else {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"username/contact_type/external_id\" for import", d.Id())
	}

	d.Set("user_name", username)
	d.Set("type", contactType)
	d.SetId(contactID)

	resourceContactRead(d, m)

	// Check that the ID was not re-set to an empty string. If it was, then the contact was not found
	if d.Id() == "" {
		return nil, fmt.Errorf("contact %s not found.", idAttr)
	}

	// In a normal read we do not set "value" as we can't compare it to the value returned from the API
	// we use computed_value for that. But in this case we need to initialize "value" with what the API
	// returned
	d.Set("value", d.Get("computed_value").(string))

	return []*schema.ResourceData{d}, nil
}
