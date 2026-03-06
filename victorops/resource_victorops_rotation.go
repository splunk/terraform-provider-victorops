package victorops

// victorops_rotation manages a rotation group (and its shifts) under a VictorOps team.
//
// API endpoints used:
//   POST   /api-public/v1/teams/{team}/rotations          — create rotation group
//   GET    /api-public/v1/teams/{team}/rotations          — list (used for read/import)
//   DELETE /api-public/v1/teams/{team}/rotations/{groupId} — delete rotation group
//
// There is no PUT endpoint for rotation groups, so all fields are ForceNew.
// The rotation group slug (rtg-xxx) is exposed as a computed attribute for use
// in victorops_escalation_policy step entries.

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// --- API request/response types ---

type rotationGroupCreatePayload struct {
	Label  string                       `json:"label"`
	Shifts []rotationShiftCreatePayload `json:"shifts,omitempty"`
}

type rotationShiftCreatePayload struct {
	Label     string        `json:"label"`
	Timezone  string        `json:"timezone"`
	Start     int64         `json:"start"`
	Duration  int           `json:"duration"`
	Shifttype string        `json:"shifttype"`
	Mask      *rotationMask `json:"mask"`
	Mask2     *rotationMask `json:"mask2,omitempty"`
	Mask3     *rotationMask `json:"mask3,omitempty"`
	Usernames []string      `json:"usernames,omitempty"`
}

type rotationMask struct {
	Day  rotationMaskDay `json:"day"`
	Time []maskTimeRange `json:"time,omitempty"`
}

type rotationMaskDay struct {
	Su bool `json:"su"`
	M  bool `json:"m"`
	T  bool `json:"t"`
	W  bool `json:"w"`
	Th bool `json:"th"`
	F  bool `json:"f"`
	Sa bool `json:"sa"`
}

type maskTimeRange struct {
	Start maskTime `json:"start"`
	End   maskTime `json:"end"`
}

type maskTime struct {
	Hour   float64 `json:"hour"`
	Minute float64 `json:"minute"`
}

// rotationGroupResponse is returned by POST.
type rotationGroupResponse struct {
	OrgSlug  string `json:"orgslug"`
	TeamSlug string `json:"teamslug"`
	ID       int    `json:"id"`
	Label    string `json:"label"`
}

// rotationGroupItem is one item from GET v1 response.
// Note: the API spec declares groupId as string but actually returns an integer.
type rotationGroupItem struct {
	TeamSlug string `json:"teamSlug"`
	Slug     string `json:"slug"`
	Label    string `json:"label"`
	GroupID  int    `json:"groupId"`
}

type rotationGroupListResponse struct {
	RotationGroups []rotationGroupItem `json:"rotationGroups"`
}

// --- Schema helpers ---

func maskSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"sunday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"monday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"tuesday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"wednesday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"thursday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"friday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"saturday": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"time_range": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "Time ranges when on-call is active. Required by the API — use start 0:00/end 23:59 for 24x7 coverage.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_hour": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"start_minute": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 59),
						},
						"end_hour": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"end_minute": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
		},
	}
}

// --- Resource definition ---

func resourceRotation() *schema.Resource {
	return &schema.Resource{
		Create: resourceRotationCreate,
		Read:   resourceRotationRead,
		Delete: resourceRotationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRotationImport,
		},

		Schema: map[string]*schema.Schema{
			"team_slug": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The team slug under which to create the rotation group (e.g. \"team-abc123\").",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The display name of the rotation group.",
			},
			// slug is the rtg-xxx identifier returned by the API; used in escalation policy entries.
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique slug of the rotation group (e.g. \"rtg-abc123\"). Use this in escalation policy step entries.",
			},
			"shift": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "One or more shifts to create within the rotation group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:     schema.TypeString,
							Required: true,
						},
						"timezone": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "IANA timezone name (e.g. \"America/Denver\", \"Europe/London\").",
						},
						"start": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Shift start time as Unix epoch milliseconds.",
						},
						"duration": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Rotation period in days. Must be 90 or less.",
							ValidateFunc: validation.IntBetween(1, 90),
						},
						"shifttype": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Shift type: \"std\" (Standard), \"fts\" (Follow The Sun), \"cstm\" (Custom), \"pho\" (Period Hand Off).",
							ValidateFunc: validation.StringInSlice(
								[]string{"std", "fts", "cstm", "pho"}, false),
						},
						"usernames": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Ordered list of usernames in the rotation.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"mask": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Primary active-period mask defining which days and times are on-call.",
							Elem:        maskSchema(),
						},
						"mask2": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Secondary mask (used with \"fts\" and \"cstm\" shift types).",
							Elem:        maskSchema(),
						},
						"mask3": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Tertiary mask (used with \"cstm\" shift type).",
							Elem:        maskSchema(),
						},
					},
				},
			},
		},
	}
}

// --- CRUD functions ---

func resourceRotationCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	teamSlug := d.Get("team_slug").(string)

	payload := rotationGroupCreatePayload{
		Label:  d.Get("label").(string),
		Shifts: expandRotationShifts(d.Get("shift").([]interface{})),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializing rotation group: %s", err)
	}

	resp, err := config.makeAPICall("POST", fmt.Sprintf("v1/teams/%s/rotations", teamSlug), body)
	if err != nil {
		return fmt.Errorf("error creating rotation group: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to create rotation group (%d): %s", resp.StatusCode, resp.ResponseBody)
	}

	var created rotationGroupResponse
	if err := json.Unmarshal([]byte(resp.ResponseBody), &created); err != nil {
		return fmt.Errorf("error parsing create rotation response: %s", err)
	}
	if created.ID == 0 {
		return fmt.Errorf("VictorOps returned a rotation group with no ID")
	}

	// Resource ID encodes both pieces needed for Read and Delete.
	d.SetId(fmt.Sprintf("%s/%d", teamSlug, created.ID))
	return resourceRotationRead(d, m)
}

func resourceRotationRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	teamSlug, groupIDStr, err := parseRotationID(d.Id())
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		return fmt.Errorf("invalid groupId %q in resource ID: %s", groupIDStr, err)
	}

	resp, err := config.makeAPICall("GET", fmt.Sprintf("v1/teams/%s/rotations", teamSlug), nil)
	if err != nil {
		return fmt.Errorf("error reading rotation groups: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to list rotation groups (%d): %s", resp.StatusCode, resp.ResponseBody)
	}

	var list rotationGroupListResponse
	if err := json.Unmarshal([]byte(resp.ResponseBody), &list); err != nil {
		return fmt.Errorf("error parsing rotation group list: %s", err)
	}

	for _, rg := range list.RotationGroups {
		if rg.GroupID == groupID {
			if err := d.Set("team_slug", rg.TeamSlug); err != nil {
				return err
			}
			if err := d.Set("label", rg.Label); err != nil {
				return err
			}
			if err := d.Set("slug", rg.Slug); err != nil {
				return err
			}
			return nil
		}
	}

	// Rotation no longer exists — remove from state.
	d.SetId("")
	return nil
}

func resourceRotationDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	teamSlug, groupIDStr, err := parseRotationID(d.Id())
	if err != nil {
		return err
	}

	resp, err := config.makeAPICall("DELETE",
		fmt.Sprintf("v1/teams/%s/rotations/%s", teamSlug, groupIDStr), nil)
	if err != nil {
		return fmt.Errorf("error deleting rotation group: %s", err)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete rotation group (%d): %s", resp.StatusCode, resp.ResponseBody)
	}

	return nil
}

func resourceRotationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Import ID format: team_slug/groupId  (e.g. "team-abc123/42")
	if _, _, err := parseRotationID(d.Id()); err != nil {
		return nil, fmt.Errorf("invalid rotation import ID %q: expected \"team_slug/groupId\"", d.Id())
	}
	if err := resourceRotationRead(d, m); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// --- Helpers ---

func parseRotationID(id string) (teamSlug, groupIDStr string, err error) {
	parts := splitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid rotation resource ID %q: expected \"team_slug/groupId\"", id)
	}
	return parts[0], parts[1], nil
}

func splitN(s, sep string, n int) []string {
	result := []string{}
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	result = append(result, s)
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func expandRotationShifts(raw []interface{}) []rotationShiftCreatePayload {
	shifts := make([]rotationShiftCreatePayload, 0, len(raw))
	for _, v := range raw {
		m := v.(map[string]interface{})

		shift := rotationShiftCreatePayload{
			Label:     m["label"].(string),
			Timezone:  m["timezone"].(string),
			Start:     int64(m["start"].(int)),
			Duration:  m["duration"].(int),
			Shifttype: m["shifttype"].(string),
			Mask:      expandMask(m["mask"].([]interface{})),
		}

		if v, ok := m["mask2"]; ok {
			if masks := v.([]interface{}); len(masks) > 0 {
				shift.Mask2 = expandMask(masks)
			}
		}
		if v, ok := m["mask3"]; ok {
			if masks := v.([]interface{}); len(masks) > 0 {
				shift.Mask3 = expandMask(masks)
			}
		}

		if rawUsers, ok := m["usernames"]; ok {
			for _, u := range rawUsers.([]interface{}) {
				shift.Usernames = append(shift.Usernames, u.(string))
			}
		}

		shifts = append(shifts, shift)
	}
	return shifts
}

func expandMask(raw []interface{}) *rotationMask {
	if len(raw) == 0 {
		return nil
	}
	m := raw[0].(map[string]interface{})

	mask := &rotationMask{
		Day: rotationMaskDay{
			Su: m["sunday"].(bool),
			M:  m["monday"].(bool),
			T:  m["tuesday"].(bool),
			W:  m["wednesday"].(bool),
			Th: m["thursday"].(bool),
			F:  m["friday"].(bool),
			Sa: m["saturday"].(bool),
		},
	}

	if rawRanges, ok := m["time_range"]; ok {
		for _, rv := range rawRanges.([]interface{}) {
			r := rv.(map[string]interface{})
			mask.Time = append(mask.Time, maskTimeRange{
				Start: maskTime{
					Hour:   float64(r["start_hour"].(int)),
					Minute: float64(r["start_minute"].(int)),
				},
				End: maskTime{
					Hour:   float64(r["end_hour"].(int)),
					Minute: float64(r["end_minute"].(int)),
				},
			})
		}
	}

	return mask
}
