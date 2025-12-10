package victorops

import (
	"encoding/json"
	"fmt"
)

// TeamOncallSchedule represents a team's on-call schedule (spec lines 4707-4715)
type TeamOncallSchedule struct {
	Team      *TeamInfo        `json:"team,omitempty"` // Object, not string
	Schedules []PolicySchedule `json:"schedules,omitempty"`
}

// TeamInfo represents team information in schedule responses (spec lines 4608-4622)
type TeamInfo struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// User represents a user object (spec lines 4602-4607)
type User struct {
	Username string `json:"username,omitempty"`
}

// EscalationPolicySummary represents policy summary (spec lines 5215-5235)
type EscalationPolicySummary struct {
	Name    string `json:"name,omitempty"`
	Slug    string `json:"slug,omitempty"`
	SelfURL string `json:"_selfUrl,omitempty"`
}

// OnCallRoll represents a roll in the schedule (spec lines 4637-4653)
type OnCallRoll struct {
	Start      string `json:"start,omitempty"`      // ISO8601 string
	End        string `json:"end,omitempty"`        // ISO8601 string
	OnCallUser *User  `json:"onCallUser,omitempty"` // Object, not string
	IsRoll     bool   `json:"isRoll,omitempty"`
}

// OnCallOverride represents an override (spec lines 4655-4667)
type OnCallOverride struct {
	OrigOnCallUser     *User  `json:"origOnCallUser,omitempty"`
	OverrideOnCallUser *User  `json:"overrideOnCallUser,omitempty"`
	Start              string `json:"start,omitempty"` // ISO8601
	End                string `json:"end,omitempty"`   // ISO8601
}

// OnCallEntry represents an on-call entry (spec lines 4669-4691)
type OnCallEntry struct {
	OnCallUser         *User        `json:"onCallUser,omitempty"` // Object, not string
	OverrideOnCallUser *User        `json:"overrideOnCallUser,omitempty"`
	OnCallType         string       `json:"onCallType,omitempty"`
	RotationName       string       `json:"rotationName,omitempty"`
	ShiftName          string       `json:"shiftName,omitempty"`
	ShiftRoll          string       `json:"shiftRoll,omitempty"` // ISO8601
	Rolls              []OnCallRoll `json:"rolls,omitempty"`
}

// PolicySchedule represents a policy schedule (spec lines 4693-4705)
type PolicySchedule struct {
	Policy    *EscalationPolicySummary `json:"policy,omitempty"`
	Schedule  []OnCallEntry            `json:"schedule,omitempty"`
	Overrides []OnCallOverride         `json:"overrides,omitempty"`
}

// Rotation represents a rotation in a team (v2 API)
type Rotation struct {
	Label                  string         `json:"label,omitempty"`
	TotalMembersInRotation int            `json:"totalMembersInRotation,omitempty"`
	Shifts                 []ShiftDetails `json:"shifts,omitempty"`
}

// ShiftDetails represents details of a shift
type ShiftDetails struct {
	Label        string         `json:"label,omitempty"`
	Duration     int            `json:"duration,omitempty"`
	Current      *OnCallPeriod  `json:"current,omitempty"`
	Next         *OnCallPeriod  `json:"next,omitempty"`
	Periods      []OnCallPeriod `json:"periods,omitempty"`
	ShiftMembers []ShiftMember  `json:"shiftMembers,omitempty"`
	ShiftType    string         `json:"shifttype,omitempty"`
	Start        string         `json:"start,omitempty"`
	Timezone     string         `json:"timezone,omitempty"`
	Mask         *RotationMask  `json:"mask,omitempty"`
	Mask2        *RotationMask  `json:"mask2,omitempty"`
	Mask3        *RotationMask  `json:"mask3,omitempty"`
}

// ShiftMember represents a member in a shift
type ShiftMember struct {
	Slug     string `json:"slug,omitempty"`
	Username string `json:"username,omitempty"`
}

// OnCallPeriod represents an on-call time period
type OnCallPeriod struct {
	Start    string `json:"start,omitempty"`
	End      string `json:"end,omitempty"`
	Username string `json:"username,omitempty"`
}

// RotationMask defines days and time ranges for on-call periods
type RotationMask struct {
	Day  *MaskDays       `json:"day,omitempty"`
	Time []MaskTimeRange `json:"time,omitempty"`
}

// MaskDays represents which days of the week the rotation is active
type MaskDays struct {
	Su bool `json:"su,omitempty"`
	M  bool `json:"m,omitempty"`
	T  bool `json:"t,omitempty"`
	W  bool `json:"w,omitempty"`
	Th bool `json:"th,omitempty"`
	F  bool `json:"f,omitempty"`
	Sa bool `json:"sa,omitempty"`
}

// MaskTimeRange represents a time range within a day
type MaskTimeRange struct {
	Start *MaskTime `json:"start,omitempty"`
	End   *MaskTime `json:"end,omitempty"`
}

// MaskTime represents hour and minute
type MaskTime struct {
	Hour   int `json:"hour,omitempty"`
	Minute int `json:"minute,omitempty"`
}

// TeamAdmin represents a team admin
type TeamAdmin struct {
	Username string `json:"username,omitempty"`
}

// RoutingKeyInfo represents a routing key with target info
type RoutingKeyInfo struct {
	RoutingKey string             `json:"routingKey,omitempty"`
	Targets    []RoutingKeyTarget `json:"targets,omitempty"`
	IsDefault  bool               `json:"isDefault,omitempty"`
}

// RoutingKeyTarget represents a target for a routing key
type RoutingKeyTarget struct {
	PolicySlug string `json:"policySlug,omitempty"`
	PolicyName string `json:"policyName,omitempty"`
}

// UserInfo represents user information
type UserInfo struct {
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Admin     bool   `json:"admin,omitempty"`
}

// UserDevice represents a user device
type UserDevice struct {
	ExtID      string `json:"extId,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
	Label      string `json:"label,omitempty"`
}

// GetTeamOncallSchedule gets a team's on-call schedule
func (c *APIClient) GetTeamOncallSchedule(teamID string, daysForward, daysSkip int) (*TeamOncallSchedule, error) {
	path := fmt.Sprintf("/api-public/v2/team/%s/oncall/schedule?daysForward=%d&daysSkip=%d", teamID, daysForward, daysSkip)
	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var schedule TeamOncallSchedule
	if err := json.Unmarshal(respBody, &schedule); err != nil {
		return nil, err
	}

	return &schedule, nil
}

// GetUserOncallSchedule gets a user's on-call schedule
func (c *APIClient) GetUserOncallSchedule(username string, daysForward, daysSkip int) ([]TeamOncallSchedule, error) {
	path := fmt.Sprintf("/api-public/v2/user/%s/oncall/schedule?daysForward=%d&daysSkip=%d", username, daysForward, daysSkip)
	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		TeamSchedules []TeamOncallSchedule `json:"teamSchedules"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response.TeamSchedules, nil
}

// GetTeamRotations gets a team's rotations
func (c *APIClient) GetTeamRotations(teamID string) ([]Rotation, error) {
	path := fmt.Sprintf("/api-public/v2/team/%s/rotations", teamID)
	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		Rotations []Rotation `json:"rotations"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response.Rotations, nil
}

// GetTeamAdmins gets a team's admins
func (c *APIClient) GetTeamAdmins(teamID string) ([]TeamAdmin, error) {
	path := fmt.Sprintf("/api-public/v1/team/%s/admins", teamID)
	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		TeamAdmins []TeamAdmin `json:"teamAdmins"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response.TeamAdmins, nil
}

// GetRoutingKeys gets all routing keys
func (c *APIClient) GetRoutingKeys() ([]RoutingKeyInfo, error) {
	path := "/api-public/v1/org/routing-keys"
	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		RoutingKeys []RoutingKeyInfo `json:"routingKeys"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response.RoutingKeys, nil
}

// GetUsers gets all users, optionally filtered by email
func (c *APIClient) GetUsers(email string) ([]UserInfo, error) {
	path := "/api-public/v2/user"
	if email != "" {
		path = fmt.Sprintf("%s?email=%s", path, email)
	}

	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		Users []UserInfo `json:"users"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response.Users, nil
}

// GetUserDevices gets a user's devices
func (c *APIClient) GetUserDevices(username string) ([]UserDevice, error) {
	path := fmt.Sprintf("/api-public/v1/user/%s/contact-methods/devices", username)
	respBody, statusCode, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	// Try to unmarshal as a direct array first
	var devices []UserDevice
	if err := json.Unmarshal(respBody, &devices); err != nil {
		// If that fails, try wrapped response
		var response struct {
			Devices []UserDevice `json:"devices"`
		}
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse devices response: %s", string(respBody))
		}
		return response.Devices, nil
	}

	return devices, nil
}
