package victorops

// victorops_scheduled_override manages a temporary on-call override in VictorOps.
//
// The VictorOps API endpoint is:
//   POST   /api-public/v1/overrides          — create
//   GET    /api-public/v1/overrides          — list (used for read)
//   DELETE /api-public/v1/overrides/{publicId} — delete
//
// Because the SDK does not cover this endpoint, we call it directly using
// Config.makeAPICall, which signs requests with the same API credentials.

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// scheduledOverrideRequest is the JSON body sent to POST /v1/overrides.
type scheduledOverrideRequest struct {
	Username   string `json:"username"`
	Start      string `json:"start"`
	End        string `json:"end"`
	PolicySlug string `json:"policySlug,omitempty"`
}

// scheduledOverride represents a single override returned in the API list.
type scheduledOverride struct {
	PublicID   string `json:"publicId"`
	Username   string `json:"username"`
	Start      string `json:"start"`
	End        string `json:"end"`
	PolicySlug string `json:"policySlug"`
	OrigUser   string `json:"origOnCallUser"`
}

// scheduledOverrideListResponse is the top-level list response from GET /v1/overrides.
type scheduledOverrideListResponse struct {
	Overrides []scheduledOverride `json:"overrides"`
}

func resourceScheduledOverride() *schema.Resource {
	return &schema.Resource{
		Create: resourceScheduledOverrideCreate,
		Read:   resourceScheduledOverrideRead,
		Delete: resourceScheduledOverrideDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// The user who will take over on-call duty.
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Username of the user taking over on-call duty.",
			},
			// RFC3339 timestamps for the override window.
			"start": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Start of the override window in RFC3339 format (e.g. \"2024-01-01T08:00:00Z\").",
			},
			"end": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "End of the override window in RFC3339 format.",
			},
			// Optionally scope the override to a specific escalation policy.
			"policy_slug": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Slug of the escalation policy to restrict the override to. If omitted, overrides all policies.",
			},
			// The original on-call user being replaced — returned by the API and stored in state.
			"original_on_call_user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The username of the original on-call user who was replaced.",
			},
		},
	}
}

func resourceScheduledOverrideCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	req := scheduledOverrideRequest{
		Username:   d.Get("username").(string),
		Start:      d.Get("start").(string),
		End:        d.Get("end").(string),
		PolicySlug: d.Get("policy_slug").(string),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := config.makeAPICall("POST", "v1/overrides", body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to create scheduled override (%d): %s", resp.StatusCode, resp.ResponseBody)
	}

	// The create response returns the new override object; parse it to get the publicId.
	var override scheduledOverride
	if err := json.Unmarshal([]byte(resp.ResponseBody), &override); err != nil {
		return fmt.Errorf("error parsing scheduled override response: %s", err)
	}

	if override.PublicID == "" {
		return fmt.Errorf("VictorOps returned a scheduled override with no publicId")
	}

	d.SetId(override.PublicID)
	return resourceScheduledOverrideRead(d, m)
}

func resourceScheduledOverrideRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	publicID := d.Id()

	// The VictorOps API only exposes a list endpoint, so we fetch all overrides and find ours.
	resp, err := config.makeAPICall("GET", "v1/overrides", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error listing scheduled overrides (%d): %s", resp.StatusCode, resp.ResponseBody)
	}

	var listResp scheduledOverrideListResponse
	if err := json.Unmarshal([]byte(resp.ResponseBody), &listResp); err != nil {
		return fmt.Errorf("error parsing scheduled overrides list: %s", err)
	}

	for _, o := range listResp.Overrides {
		if o.PublicID == publicID {
			if err := d.Set("username", o.Username); err != nil {
				return err
			}
			if err := d.Set("start", o.Start); err != nil {
				return err
			}
			if err := d.Set("end", o.End); err != nil {
				return err
			}
			if err := d.Set("policy_slug", o.PolicySlug); err != nil {
				return err
			}
			if err := d.Set("original_on_call_user", o.OrigUser); err != nil {
				return err
			}
			return nil
		}
	}

	// Override no longer exists — remove it from state.
	d.SetId("")
	return nil
}

func resourceScheduledOverrideDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	resp, err := config.makeAPICall("DELETE", "v1/overrides/"+d.Id(), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to delete scheduled override (%d): %s", resp.StatusCode, resp.ResponseBody)
	}

	return nil
}
