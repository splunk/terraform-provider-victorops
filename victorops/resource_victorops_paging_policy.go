package victorops

// victorops_paging_policy manages the per-user paging policy in VictorOps.
//
// The VictorOps API endpoint is:
//   GET /api-public/v1/profile/{username}/policies — read
//   PUT /api-public/v1/profile/{username}/policies — create / update (replaces entire policy)
//
// Because the API uses a full-replace PUT, both Create and Update call the same function.
// The resource ID is set to the username.

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// pagingPolicyContact maps to a single contact entry in a paging policy step.
type pagingPolicyContact struct {
	Type      string `json:"type"`                // "phone", "email", or "push"
	ContactID int    `json:"contactId,omitempty"` // required for phone/email, omitted for push
}

// pagingPolicyStep maps to one step in the paging policy.
type pagingPolicyStep struct {
	Timeout  int                   `json:"timeout"`
	Contacts []pagingPolicyContact `json:"contacts"`
}

// pagingPolicyRequest is the JSON body sent to PUT /v1/profile/{username}/policies.
type pagingPolicyRequest struct {
	Steps []pagingPolicyStep `json:"steps"`
}

func resourcePagingPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagingPolicyCreate,
		Read:   resourcePagingPolicyRead,
		Update: resourcePagingPolicyUpdate,
		Delete: resourcePagingPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VictorOps username whose paging policy this resource manages.",
			},
			// Each step is a block that describes one escalation step.
			"step": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Ordered list of paging policy steps. Each step is tried after the previous times out.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Minutes to wait before escalating to the next step. Use 0 for immediate.",
						},
						// Each step may notify via one or more contact methods.
						"contact": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "One or more contact methods to use in this step.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Contact method type: \"phone\", \"email\", or \"push\".",
										ValidateFunc: validation.StringInSlice([]string{"phone", "email", "push"}, false),
									},
									// contact_id comes from the victorops_contact resource's internal_id attribute.
									"contact_id": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Internal contact ID (from victorops_contact.internal_id). Required for phone/email, omitted for push.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// buildPagingPolicyRequest converts the Terraform resource data into the API request body.
func buildPagingPolicyRequest(d *schema.ResourceData) (*pagingPolicyRequest, error) {
	rawSteps := d.Get("step").([]interface{})
	steps := make([]pagingPolicyStep, 0, len(rawSteps))

	for _, rawStep := range rawSteps {
		stepMap := rawStep.(map[string]interface{})
		rawContacts := stepMap["contact"].([]interface{})
		contacts := make([]pagingPolicyContact, 0, len(rawContacts))

		for _, rawContact := range rawContacts {
			cm := rawContact.(map[string]interface{})
			c := pagingPolicyContact{
				Type:      cm["type"].(string),
				ContactID: cm["contact_id"].(int),
			}
			contacts = append(contacts, c)
		}

		steps = append(steps, pagingPolicyStep{
			Timeout:  stepMap["timeout"].(int),
			Contacts: contacts,
		})
	}

	return &pagingPolicyRequest{Steps: steps}, nil
}

func resourcePagingPolicyCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Get("username").(string)

	req, err := buildPagingPolicyRequest(d)
	if err != nil {
		return err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := config.makeAPICall("PUT", "v1/profile/"+username+"/policies", body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to create paging policy for %s (%d): %s", username, resp.StatusCode, resp.ResponseBody)
	}

	d.SetId(username)
	return resourcePagingPolicyRead(d, m)
}

func resourcePagingPolicyRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Id()

	resp, err := config.makeAPICall("GET", "v1/profile/"+username+"/policies", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error reading paging policy for %s (%d): %s", username, resp.StatusCode, resp.ResponseBody)
	}

	// The API returns the full policy; unmarshal into the same request struct since the shape is identical.
	var policy pagingPolicyRequest
	if err := json.Unmarshal([]byte(resp.ResponseBody), &policy); err != nil {
		return fmt.Errorf("error parsing paging policy response: %s", err)
	}

	// Convert the API response back into a list of maps for Terraform state.
	steps := make([]map[string]interface{}, 0, len(policy.Steps))
	for _, s := range policy.Steps {
		contacts := make([]map[string]interface{}, 0, len(s.Contacts))
		for _, c := range s.Contacts {
			contacts = append(contacts, map[string]interface{}{
				"type":       c.Type,
				"contact_id": c.ContactID,
			})
		}
		steps = append(steps, map[string]interface{}{
			"timeout": s.Timeout,
			"contact": contacts,
		})
	}

	if err := d.Set("step", steps); err != nil {
		return err
	}
	if err := d.Set("username", username); err != nil {
		return err
	}

	return nil
}

func resourcePagingPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	// PUT is a full replace, so update is identical to create.
	return resourcePagingPolicyCreate(d, m)
}

// resourcePagingPolicyDelete resets the user's paging policy to an empty state.
// VictorOps does not have a DELETE endpoint for paging policies; we overwrite with an empty step list instead.
func resourcePagingPolicyDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	username := d.Id()

	empty := pagingPolicyRequest{Steps: []pagingPolicyStep{}}
	body, err := json.Marshal(empty)
	if err != nil {
		return err
	}

	resp, err := config.makeAPICall("PUT", "v1/profile/"+username+"/policies", body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to clear paging policy for %s (%d): %s", username, resp.StatusCode, resp.ResponseBody)
	}

	return nil
}
