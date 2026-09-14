package victorops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const rotationRequestInterval = 500 * time.Millisecond

type rotationClient struct {
	apiID       string
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	minInterval time.Duration
	mutex       sync.Mutex
	lastRequest time.Time
}

type rotationGroupPayload struct {
	Label  string                 `json:"label"`
	Shifts []rotationShiftPayload `json:"shifts,omitempty"`
}

type rotationGroupUpdatePayload struct {
	Label string `json:"label"`
}

type rotationGroupResponse struct {
	ID       int                     `json:"id"`
	Label    string                  `json:"label"`
	Slug     string                  `json:"slug"`
	TeamSlug string                  `json:"teamslug"`
	Shifts   []rotationShiftResponse `json:"shifts"`
}

type rotationGroupList struct {
	RotationGroups []rotationGroupSummary `json:"rotationGroups"`
}

type rotationGroupSummary struct {
	GroupID flexibleInt `json:"groupId"`
	Label   string      `json:"label"`
	Slug    string      `json:"slug"`
}

type flexibleInt int

func (i *flexibleInt) UnmarshalJSON(value []byte) error {
	text := strings.Trim(string(value), `"`)
	parsed, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("invalid integer %q: %w", text, err)
	}
	*i = flexibleInt(parsed)
	return nil
}

type rotationShiftPayload struct {
	Label     string        `json:"label"`
	Timezone  string        `json:"timezone"`
	Start     int64         `json:"start"`
	Duration  int           `json:"duration"`
	ShiftType string        `json:"shifttype"`
	Mask      rotationMask  `json:"mask"`
	Mask2     *rotationMask `json:"mask2,omitempty"`
	Mask3     *rotationMask `json:"mask3,omitempty"`
	Usernames []string      `json:"usernames,omitempty"`
}

type rotationShiftResponse struct {
	rotationShiftPayload
	RotationID   int                   `json:"rot_id"`
	GroupID      int                   `json:"group_id"`
	Usernames    []string              `json:"usernames"`
	ShiftMembers []rotationShiftMember `json:"shiftMembers"`
}

type rotationShiftMember struct {
	Slug     string `json:"slug"`
	Username string `json:"username"`
}

type rotationMask struct {
	Day  rotationDays        `json:"day"`
	Time []rotationTimeRange `json:"time"`
}

type rotationDays struct {
	Sunday    bool `json:"su"`
	Monday    bool `json:"m"`
	Tuesday   bool `json:"t"`
	Wednesday bool `json:"w"`
	Thursday  bool `json:"th"`
	Friday    bool `json:"f"`
	Saturday  bool `json:"sa"`
}

type rotationTimeRange struct {
	Start rotationTime `json:"start"`
	End   rotationTime `json:"end"`
}

type rotationTime struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

type rotationMemberAddPayload struct {
	Username string `json:"username"`
	Position *int   `json:"position,omitempty"`
}

type rotationMemberPositionPayload struct {
	Username string `json:"username"`
	Position int    `json:"position"`
}

type rotationMemberRemovePayload struct {
	Username string `json:"username"`
}

func newRotationClient(apiID, apiKey, baseURL string) *rotationClient {
	return &rotationClient{
		apiID:       apiID,
		apiKey:      apiKey,
		baseURL:     strings.TrimRight(baseURL, "/"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		minInterval: rotationRequestInterval,
	}
}

func (c *rotationClient) request(method, endpoint string, payload interface{}, target interface{}) (int, []byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(encoded)
	}

	c.waitForRequestSlot()
	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("X-VO-Api-Id", c.apiID)
	req.Header.Set("X-VO-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, nil, err
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 && target != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, target); err != nil {
			return response.StatusCode, responseBody, fmt.Errorf("failed to decode VictorOps response: %w", err)
		}
	}
	return response.StatusCode, responseBody, nil
}

func (c *rotationClient) waitForRequestSlot() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if delay := c.minInterval - time.Since(c.lastRequest); delay > 0 {
		time.Sleep(delay)
	}
	c.lastRequest = time.Now()
}

func rotationTeamPath(team string) string {
	return "/api-public/v1/teams/" + url.PathEscape(team) + "/rotations"
}

func rotationGroupPath(team string, groupID int) string {
	return rotationTeamPath(team) + "/" + strconv.Itoa(groupID)
}

func rotationShiftPath(team string, groupID, shiftID int) string {
	return rotationGroupPath(team, groupID) + "/" + strconv.Itoa(shiftID)
}

func (c *rotationClient) createGroup(team string, payload rotationGroupPayload) (*rotationGroupResponse, int, []byte, error) {
	group := &rotationGroupResponse{}
	status, body, err := c.request(http.MethodPost, rotationTeamPath(team), payload, group)
	return group, status, body, err
}

func (c *rotationClient) getGroup(team string, groupID int) (*rotationGroupResponse, int, []byte, error) {
	group := &rotationGroupResponse{}
	status, body, err := c.request(http.MethodGet, rotationGroupPath(team, groupID), nil, group)
	return group, status, body, err
}

func (c *rotationClient) listGroups(team string) (*rotationGroupList, int, []byte, error) {
	groups := &rotationGroupList{}
	status, body, err := c.request(http.MethodGet, rotationTeamPath(team), nil, groups)
	return groups, status, body, err
}

func (c *rotationClient) updateGroup(team string, groupID int, label string) (int, []byte, error) {
	return c.request(http.MethodPut, rotationGroupPath(team, groupID), rotationGroupUpdatePayload{Label: label}, nil)
}

func (c *rotationClient) deleteGroup(team string, groupID int) (int, []byte, error) {
	return c.request(http.MethodDelete, rotationGroupPath(team, groupID), nil, nil)
}

func (c *rotationClient) createShift(team string, groupID int, payload rotationShiftPayload) (*rotationShiftResponse, int, []byte, error) {
	shift := &rotationShiftResponse{}
	status, body, err := c.request(http.MethodPost, rotationGroupPath(team, groupID), payload, shift)
	return shift, status, body, err
}

func (c *rotationClient) updateShift(team string, groupID, shiftID int, payload rotationShiftPayload) (int, []byte, error) {
	payload.Usernames = nil
	return c.request(http.MethodPut, rotationShiftPath(team, groupID, shiftID), payload, nil)
}

func (c *rotationClient) deleteShift(team string, groupID, shiftID int) (int, []byte, error) {
	return c.request(http.MethodDelete, rotationShiftPath(team, groupID, shiftID), nil, nil)
}

func (c *rotationClient) addMember(team string, groupID, shiftID int, username string) (int, []byte, error) {
	endpoint := rotationShiftPath(team, groupID, shiftID) + "/members"
	return c.request(http.MethodPost, endpoint, rotationMemberAddPayload{Username: username}, nil)
}

func (c *rotationClient) positionMember(team string, groupID, shiftID int, username string, position int) (int, []byte, error) {
	endpoint := rotationShiftPath(team, groupID, shiftID) + "/members"
	payload := rotationMemberPositionPayload{Username: username, Position: position}
	return c.request(http.MethodPut, endpoint, payload, nil)
}

func (c *rotationClient) removeMember(team string, groupID, shiftID int, username string) (int, []byte, error) {
	endpoint := rotationShiftPath(team, groupID, shiftID) + "/members"
	return c.request(http.MethodDelete, endpoint, rotationMemberRemovePayload{Username: username}, nil)
}
