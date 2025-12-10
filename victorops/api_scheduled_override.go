package victorops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ScheduledOverride represents a scheduled override in VictorOps
type ScheduledOverride struct {
	PublicID    string                 `json:"publicId,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Timezone    string                 `json:"timezone,omitempty"`
	Start       string                 `json:"start,omitempty"`
	End         string                 `json:"end,omitempty"`
	Assignments []OverrideAssignment   `json:"assignments,omitempty"`
	User        *ScheduledOverrideUser `json:"user,omitempty"`
}

// ScheduledOverrideUser represents the user part of a scheduled override
type ScheduledOverrideUser struct {
	Username string `json:"username,omitempty"`
}

// OverrideAssignment represents an assignment for a scheduled override
type OverrideAssignment struct {
	Team     string `json:"team,omitempty"`
	Policy   string `json:"policy,omitempty"`
	Assigned bool   `json:"assigned,omitempty"`
	Username string `json:"user,omitempty"`
}

// ScheduledOverrideCreateRequest is the request body for creating a scheduled override
type ScheduledOverrideCreateRequest struct {
	Username string `json:"username"`
	Timezone string `json:"timezone"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

// ScheduledOverrideResponse is the API response for scheduled override operations
type ScheduledOverrideResponse struct {
	Override *ScheduledOverride `json:"override,omitempty"`
	Schedule *ScheduledOverride `json:"schedule,omitempty"`
	SelfURL  string             `json:"_selfUrl,omitempty"`
}

// AssignmentUpdateRequest is the request body for updating an assignment
type AssignmentUpdateRequest struct {
	Username      string `json:"username"`
	AcceptOverlap bool   `json:"acceptOverlap,omitempty"`
}

// APIClient provides methods for VictorOps API calls not in go-victorops
type APIClient struct {
	BaseURL string
	APIID   string
	APIKey  string
	Client  *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL, apiID, apiKey string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		APIID:   apiID,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("X-VO-Api-Id", c.APIID)
	req.Header.Set("X-VO-Api-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

// doRequestText sends a request with text/plain content type (no body)
func (c *APIClient) doRequestText(method, path string) ([]byte, int, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("X-VO-Api-Id", c.APIID)
	req.Header.Set("X-VO-Api-Key", c.APIKey)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

// CreateScheduledOverride creates a new scheduled override
func (c *APIClient) CreateScheduledOverride(req *ScheduledOverrideCreateRequest) (*ScheduledOverride, error) {
	respBody, statusCode, err := c.doRequest("POST", "/api-public/v1/overrides", req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response ScheduledOverrideResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	if response.Schedule != nil {
		return response.Schedule, nil
	}
	return response.Override, nil
}

// GetScheduledOverride gets a scheduled override by public ID
func (c *APIClient) GetScheduledOverride(publicID string) (*ScheduledOverride, error) {
	respBody, statusCode, err := c.doRequest("GET", "/api-public/v1/overrides/"+publicID, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response ScheduledOverrideResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response.Override, nil
}

// DeleteScheduledOverride deletes a scheduled override
func (c *APIClient) DeleteScheduledOverride(publicID string) error {
	_, statusCode, err := c.doRequest("DELETE", "/api-public/v1/overrides/"+publicID, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 && statusCode != 204 {
		return fmt.Errorf("API error (%d)", statusCode)
	}

	return nil
}

// UpdateAssignment updates an assignment for a scheduled override
func (c *APIClient) UpdateAssignment(publicID, policySlug string, req *AssignmentUpdateRequest) (*OverrideAssignment, error) {
	path := fmt.Sprintf("/api-public/v1/overrides/%s/assignments/%s", publicID, policySlug)
	respBody, statusCode, err := c.doRequest("PUT", path, req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var assignment OverrideAssignment
	if err := json.Unmarshal(respBody, &assignment); err != nil {
		return nil, err
	}

	return &assignment, nil
}

// DeleteAssignment deletes an assignment for a scheduled override
func (c *APIClient) DeleteAssignment(publicID, policySlug string) error {
	path := fmt.Sprintf("/api-public/v1/overrides/%s/assignments/%s", publicID, policySlug)
	_, statusCode, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 && statusCode != 204 {
		return fmt.Errorf("API error (%d)", statusCode)
	}

	return nil
}
