package victorops

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

type rotationShiftConfig struct {
	ID      int
	Payload rotationShiftPayload
	Members []string
}

func resourceRotation() *schema.Resource {
	return &schema.Resource{
		Create: resourceRotationCreate,
		Read:   resourceRotationRead,
		Update: resourceRotationUpdate,
		Delete: resourceRotationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRotationImport,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The slug of the team that owns the rotation.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The rotation group label.",
			},
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The rotation group slug used by escalation policies.",
			},
			"shift": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "An ordered list of shifts in the rotation group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"timezone": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(TimeZoneList(), false),
						},
						"start": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The shift start time in epoch milliseconds.",
						},
						"duration": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 90),
							Description:  "The shift duration in days.",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"std", "fts", "cstm", "pho"}, false),
							Description:  "The shift type: std, fts, cstm, or pho.",
						},
						"members": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Usernames in their rotation order.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"mask":  rotationMaskSchema(true),
						"mask2": rotationMaskSchema(false),
						"mask3": rotationMaskSchema(false),
					},
				},
			},
		},
	}
}

func rotationMaskSchema(required bool) *schema.Schema {
	result := &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"days": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"su", "m", "t", "w", "th", "f", "sa"}, false),
					},
				},
				"time": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
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
		},
	}
	if required {
		result.Required = true
	} else {
		result.Optional = true
		result.MinItems = 0
	}
	return result
}

func resourceRotationCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := rotationClientForConfig(config)
	shifts, err := expandRotationShifts(d.Get("shift"))
	if err != nil {
		return err
	}

	payload := rotationGroupPayload{Label: d.Get("name").(string)}
	for _, shift := range shifts {
		shift.Payload.Usernames = shift.Members
		payload.Shifts = append(payload.Shifts, shift.Payload)
	}

	group, status, body, err := client.createGroup(d.Get("team_id").(string), payload)
	if err != nil {
		return err
	}
	if err := checkRotationStatus("create rotation", status, body); err != nil {
		return err
	}
	if group.ID == 0 {
		return fmt.Errorf("failed to create rotation: response did not contain a rotation group ID")
	}

	d.SetId(strconv.Itoa(group.ID))
	return resourceRotationRead(d, m)
}

func resourceRotationRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := rotationClientForConfig(config)
	groupID, err := parseRotationID(d.Id())
	if err != nil {
		return err
	}
	team := d.Get("team_id").(string)

	group, status, body, err := client.getGroup(team, groupID)
	if err != nil {
		return err
	}
	if status == 404 {
		d.SetId("")
		return nil
	}
	if err := checkRotationStatus("read rotation", status, body); err != nil {
		return err
	}

	if err := d.Set("name", group.Label); err != nil {
		return err
	}
	if err := d.Set("shift", flattenRotationShifts(group.Shifts)); err != nil {
		return err
	}

	slug := group.Slug
	if slug == "" {
		groups, listStatus, listBody, listErr := client.listGroups(team)
		if listErr != nil {
			return listErr
		}
		if err := checkRotationStatus("list rotations", listStatus, listBody); err != nil {
			return err
		}
		for _, candidate := range groups.RotationGroups {
			if int(candidate.GroupID) == groupID {
				slug = candidate.Slug
				break
			}
		}
	}
	if slug == "" {
		return fmt.Errorf("failed to read rotation %d: the API did not return its rotation group slug", groupID)
	}
	return d.Set("slug", slug)
}

func resourceRotationUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := rotationClientForConfig(config)
	groupID, err := parseRotationID(d.Id())
	if err != nil {
		return err
	}
	team := d.Get("team_id").(string)

	group, status, body, err := client.getGroup(team, groupID)
	if err != nil {
		return err
	}
	if err := checkRotationStatus("read rotation before update", status, body); err != nil {
		return err
	}

	if d.HasChange("name") {
		status, body, err = client.updateGroup(team, groupID, d.Get("name").(string))
		if err != nil {
			return err
		}
		if err := checkRotationStatus("update rotation", status, body); err != nil {
			return err
		}
	}

	if d.HasChange("shift") {
		desired, err := expandRotationShifts(d.Get("shift"))
		if err != nil {
			return err
		}
		if err := reconcileRotationShifts(client, team, groupID, group.Shifts, desired); err != nil {
			return err
		}
	}

	return resourceRotationRead(d, m)
}

func resourceRotationDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	groupID, err := parseRotationID(d.Id())
	if err != nil {
		return err
	}
	status, body, err := rotationClientForConfig(config).deleteGroup(d.Get("team_id").(string), groupID)
	if err != nil {
		return err
	}
	if status == 404 {
		return nil
	}
	return checkRotationStatus("delete rotation", status, body)
}

func resourceRotationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("rotation import ID must be in the form team-slug/group-id")
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return nil, fmt.Errorf("rotation import group ID must be an integer: %s", parts[1])
	}
	if err := d.Set("team_id", parts[0]); err != nil {
		return nil, err
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

func rotationClientForConfig(config Config) *rotationClient {
	if config.RotationClient != nil {
		return config.RotationClient
	}
	return newRotationClient(config.APIId, config.APIKey, config.BaseURL)
}

func parseRotationID(id string) (int, error) {
	groupID, err := strconv.Atoi(id)
	if err != nil || groupID <= 0 {
		return 0, fmt.Errorf("invalid rotation group ID %q", id)
	}
	return groupID, nil
}

func checkRotationStatus(action string, status int, body []byte) error {
	if status >= 200 && status < 300 {
		return nil
	}
	return fmt.Errorf("failed to %s (%d): %s", action, status, strings.TrimSpace(string(body)))
}

func reconcileRotationShifts(client *rotationClient, team string, groupID int, current []rotationShiftResponse, desired []rotationShiftConfig) error {
	remoteByID := make(map[int]rotationShiftResponse)
	for _, shift := range current {
		remoteByID[shift.RotationID] = shift
	}
	retained := make(map[int]bool)

	for _, shift := range desired {
		remote, exists := remoteByID[shift.ID]
		if !exists || shift.ID == 0 {
			payload := shift.Payload
			payload.Usernames = nil
			created, status, body, err := client.createShift(team, groupID, payload)
			if err != nil {
				return err
			}
			if err := checkRotationStatus("create rotation shift", status, body); err != nil {
				return err
			}
			if created.RotationID == 0 {
				return fmt.Errorf("failed to create rotation shift: response did not contain a shift ID")
			}
			shift.ID = created.RotationID
			remote = *created
		} else {
			status, body, err := client.updateShift(team, groupID, shift.ID, shift.Payload)
			if err != nil {
				return err
			}
			if err := checkRotationStatus("update rotation shift", status, body); err != nil {
				return err
			}
		}
		retained[shift.ID] = true
		if err := syncRotationMembers(client, team, groupID, shift.ID, rotationMemberNames(remote), shift.Members); err != nil {
			return err
		}
	}

	for _, shift := range current {
		if retained[shift.RotationID] {
			continue
		}
		status, body, err := client.deleteShift(team, groupID, shift.RotationID)
		if err != nil {
			return err
		}
		if err := checkRotationStatus("delete rotation shift", status, body); err != nil {
			return err
		}
	}
	return nil
}

func expandRotationShifts(raw interface{}) ([]rotationShiftConfig, error) {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil, nil
	}
	result := make([]rotationShiftConfig, 0, len(items))
	for _, item := range items {
		values := item.(map[string]interface{})
		mask, err := expandRotationMask(values["mask"])
		if err != nil {
			return nil, fmt.Errorf("invalid mask for shift %q: %w", values["name"], err)
		}
		mask2, err := expandRotationMask(values["mask2"])
		if err != nil {
			return nil, fmt.Errorf("invalid mask2 for shift %q: %w", values["name"], err)
		}
		mask3, err := expandRotationMask(values["mask3"])
		if err != nil {
			return nil, fmt.Errorf("invalid mask3 for shift %q: %w", values["name"], err)
		}
		if mask == nil {
			return nil, fmt.Errorf("shift %q requires a mask", values["name"])
		}

		members := stringList(values["members"])
		if duplicate := duplicateString(members); duplicate != "" {
			return nil, fmt.Errorf("shift %q contains duplicate member %q", values["name"], duplicate)
		}
		shiftID, _ := values["id"].(int)
		result = append(result, rotationShiftConfig{
			ID: shiftID,
			Payload: rotationShiftPayload{
				Label:     values["name"].(string),
				Timezone:  values["timezone"].(string),
				Start:     int64(values["start"].(int)),
				Duration:  values["duration"].(int),
				ShiftType: values["type"].(string),
				Mask:      *mask,
				Mask2:     mask2,
				Mask3:     mask3,
			},
			Members: members,
		})
	}
	return result, nil
}

func expandRotationMask(raw interface{}) (*rotationMask, error) {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 || items[0] == nil {
		return nil, nil
	}
	values := items[0].(map[string]interface{})
	days := values["days"].(*schema.Set).List()
	mask := &rotationMask{}
	for _, rawDay := range days {
		switch rawDay.(string) {
		case "su":
			mask.Day.Sunday = true
		case "m":
			mask.Day.Monday = true
		case "t":
			mask.Day.Tuesday = true
		case "w":
			mask.Day.Wednesday = true
		case "th":
			mask.Day.Thursday = true
		case "f":
			mask.Day.Friday = true
		case "sa":
			mask.Day.Saturday = true
		default:
			return nil, fmt.Errorf("unsupported day %q", rawDay)
		}
	}
	for _, rawTime := range values["time"].([]interface{}) {
		timeValues := rawTime.(map[string]interface{})
		mask.Time = append(mask.Time, rotationTimeRange{
			Start: rotationTime{Hour: timeValues["start_hour"].(int), Minute: timeValues["start_minute"].(int)},
			End:   rotationTime{Hour: timeValues["end_hour"].(int), Minute: timeValues["end_minute"].(int)},
		})
	}
	return mask, nil
}

func flattenRotationShifts(shifts []rotationShiftResponse) []interface{} {
	result := make([]interface{}, 0, len(shifts))
	for _, shift := range shifts {
		result = append(result, map[string]interface{}{
			"id":       shift.RotationID,
			"name":     shift.Label,
			"timezone": shift.Timezone,
			"start":    int(shift.Start),
			"duration": shift.Duration,
			"type":     shift.ShiftType,
			"members":  rotationMemberNames(shift),
			"mask":     flattenRotationMask(&shift.Mask),
			"mask2":    flattenRotationMask(shift.Mask2),
			"mask3":    flattenRotationMask(shift.Mask3),
		})
	}
	return result
}

func flattenRotationMask(mask *rotationMask) []interface{} {
	if mask == nil {
		return nil
	}
	days := make([]string, 0, 7)
	for _, day := range []struct {
		name    string
		enabled bool
	}{
		{"su", mask.Day.Sunday},
		{"m", mask.Day.Monday},
		{"t", mask.Day.Tuesday},
		{"w", mask.Day.Wednesday},
		{"th", mask.Day.Thursday},
		{"f", mask.Day.Friday},
		{"sa", mask.Day.Saturday},
	} {
		if day.enabled {
			days = append(days, day.name)
		}
	}
	times := make([]interface{}, 0, len(mask.Time))
	for _, item := range mask.Time {
		times = append(times, map[string]interface{}{
			"start_hour":   item.Start.Hour,
			"start_minute": item.Start.Minute,
			"end_hour":     item.End.Hour,
			"end_minute":   item.End.Minute,
		})
	}
	return []interface{}{map[string]interface{}{"days": days, "time": times}}
}

func rotationMemberNames(shift rotationShiftResponse) []string {
	if len(shift.Usernames) > 0 {
		return append([]string(nil), shift.Usernames...)
	}
	result := make([]string, 0, len(shift.ShiftMembers))
	for _, member := range shift.ShiftMembers {
		result = append(result, member.Username)
	}
	return result
}

func stringList(raw interface{}) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}

func duplicateString(values []string) string {
	seen := make(map[string]bool)
	for _, value := range values {
		if seen[value] {
			return value
		}
		seen[value] = true
	}
	return ""
}

func syncRotationMembers(client *rotationClient, team string, groupID, shiftID int, current, desired []string) error {
	currentSet := make(map[string]bool)
	desiredSet := make(map[string]bool)
	for _, username := range current {
		currentSet[username] = true
	}
	for _, username := range desired {
		desiredSet[username] = true
		if currentSet[username] {
			continue
		}
		status, body, err := client.addMember(team, groupID, shiftID, username)
		if err != nil {
			return err
		}
		if err := checkRotationStatus("add rotation member", status, body); err != nil {
			return err
		}
	}
	for _, username := range current {
		if desiredSet[username] {
			continue
		}
		status, body, err := client.removeMember(team, groupID, shiftID, username)
		if err != nil {
			return err
		}
		if err := checkRotationStatus("remove rotation member", status, body); err != nil {
			return err
		}
	}
	for position, username := range desired {
		status, body, err := client.positionMember(team, groupID, shiftID, username, position)
		if err != nil {
			return err
		}
		if err := checkRotationStatus("reorder rotation member", status, body); err != nil {
			return err
		}
	}
	return nil
}
