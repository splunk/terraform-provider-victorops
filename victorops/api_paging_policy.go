package victorops

import (
	"encoding/json"
	"fmt"
)

// PagingPolicy represents a user's paging policy
type PagingPolicy struct {
	Steps   []PagingPolicyStep `json:"steps,omitempty"`
	SelfURL string             `json:"_selfUrl,omitempty"`
}

// Contact represents a contact for paging (spec lines 5327-5333)
type Contact struct {
	ID   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"` // "email" or "phone"
}

// PagingPolicyStep represents a step in a paging policy (spec lines 5353-5364)
type PagingPolicyStep struct {
	Index   int                `json:"index,omitempty"` // Changed from "step"
	Timeout int                `json:"timeout,omitempty"`
	Rules   []PagingPolicyRule `json:"rules,omitempty"`
	SelfURL string             `json:"_selfUrl,omitempty"`
}

// PagingPolicyRule represents a rule in a paging policy step (spec lines 5335-5343)
type PagingPolicyRule struct {
	Index   int      `json:"index,omitempty"` // Changed from "rule"
	Type    string   `json:"type,omitempty"`
	Contact *Contact `json:"contact,omitempty"` // Changed from ContactID
	SelfURL string   `json:"_selfUrl,omitempty"`
}

// PagingPolicyStepCreateRequest is the request body for creating a step (spec lines 5411-5419)
type PagingPolicyStepCreateRequest struct {
	Timeout int                          `json:"timeout"`
	Rules   []PagingPolicyRuleAddPayload `json:"rules,omitempty"`
}

// PagingPolicyRuleAddPayload represents a rule to add (spec lines 5345-5351)
type PagingPolicyRuleAddPayload struct {
	Type    string   `json:"type"`
	Contact *Contact `json:"contact,omitempty"`
}

// PagingPolicyRuleCreateRequest is the request body for creating a rule (spec lines 5421-5427)
type PagingPolicyRuleCreateRequest struct {
	Type    string   `json:"type"`
	Contact *Contact `json:"contact,omitempty"`
}

// GetUserPagingPolicy gets a user's paging policy
func (c *APIClient) GetUserPagingPolicy(username string) (*PagingPolicy, error) {
	path := fmt.Sprintf("/api-public/v1/profile/%s/policies", username)
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
		Steps []PagingPolicyStep `json:"steps"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &PagingPolicy{Steps: response.Steps}, nil
}

// CreatePagingPolicyStep creates a new step in a user's paging policy
func (c *APIClient) CreatePagingPolicyStep(username string, req *PagingPolicyStepCreateRequest) (*PagingPolicyStep, error) {
	path := fmt.Sprintf("/api-public/v1/profile/%s/policies", username)
	respBody, statusCode, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		Step PagingPolicyStep `json:"step"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response.Step, nil
}

// UpdatePagingPolicyStep updates a step in a user's paging policy
func (c *APIClient) UpdatePagingPolicyStep(username string, stepNum int, req *PagingPolicyStepCreateRequest) (*PagingPolicyStep, error) {
	path := fmt.Sprintf("/api-public/v1/profile/%s/policies/%d", username, stepNum)
	respBody, statusCode, err := c.doRequest("PUT", path, req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		Step PagingPolicyStep `json:"step"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response.Step, nil
}

// DeletePagingPolicyStep deletes a step from a user's paging policy
func (c *APIClient) DeletePagingPolicyStep(username string, stepNum int) error {
	path := fmt.Sprintf("/api-public/v1/profile/%s/policies/%d", username, stepNum)
	_, statusCode, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 && statusCode != 204 {
		return fmt.Errorf("API error (%d)", statusCode)
	}

	return nil
}

// CreatePagingPolicyRule creates a new rule in a paging policy step
func (c *APIClient) CreatePagingPolicyRule(username string, stepNum int, req *PagingPolicyRuleCreateRequest) (*PagingPolicyRule, error) {
	path := fmt.Sprintf("/api-public/v1/profile/%s/policies/%d", username, stepNum)
	respBody, statusCode, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var response struct {
		StepRule PagingPolicyRule `json:"stepRule"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response.StepRule, nil
}

// DeletePagingPolicyRule deletes a rule from a paging policy step
func (c *APIClient) DeletePagingPolicyRule(username string, stepNum, ruleNum int) error {
	path := fmt.Sprintf("/api-public/v1/profile/%s/policies/%d/%d", username, stepNum, ruleNum)
	_, statusCode, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 && statusCode != 204 {
		return fmt.Errorf("API error (%d)", statusCode)
	}

	return nil
}
