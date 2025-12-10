package victorops

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// AlertRule represents an alert rule in VictorOps (API response)
// Note: API response uses "routeKey", while request uses "routingKey"
type AlertRule struct {
	ID              int               `json:"id,omitempty"`
	AlertField      string            `json:"alertField,omitempty"`
	AlertValueMatch string            `json:"alertValueMatch,omitempty"`
	MatchType       string            `json:"matchType,omitempty"`
	Rank            int               `json:"rank,omitempty"`
	StopFlag        bool              `json:"stopFlag,omitempty"`
	LastUpdated     int64             `json:"lastUpdated,omitempty"`
	LastUpdatedBy   string            `json:"lastUpdatedBy,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	RouteKey        string            `json:"routeKey,omitempty"` // API response field (not routingKey)
	Annotations     []AlertAnnotation `json:"annotations,omitempty"`
}

// AlertAnnotation represents an annotation on an alert rule
type AlertAnnotation struct {
	ID             int    `json:"id,omitempty"`
	AnnotationType string `json:"annotationType,omitempty"`
	FieldName      string `json:"fieldName,omitempty"`
	FieldValue     string `json:"fieldValue,omitempty"`
	Flags          int    `json:"flags,omitempty"`
}

// AlertRuleCreateRequest is the request body for creating an alert rule
type AlertRuleCreateRequest struct {
	AlertField      string            `json:"alertField"`
	AlertValueMatch string            `json:"alertValueMatch"`
	MatchType       string            `json:"matchType"`
	Rank            int               `json:"rank"`
	StopFlag        bool              `json:"stopFlag"`
	Notes           string            `json:"notes,omitempty"`
	RoutingKey      string            `json:"routingKey"`
	Annotations     []AlertAnnotation `json:"annotations"`
}

// ListAlertRules gets all alert rules
func (c *APIClient) ListAlertRules() ([]AlertRule, error) {
	respBody, statusCode, err := c.doRequest("GET", "/api-public/v1/alertRules", nil)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var rules []AlertRule
	if err := json.Unmarshal(respBody, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// GetAlertRule gets an alert rule by ID
func (c *APIClient) GetAlertRule(ruleID int) (*AlertRule, error) {
	rules, err := c.ListAlertRules()
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if rule.ID == ruleID {
			return &rule, nil
		}
	}

	return nil, nil
}

// CreateAlertRule creates a new alert rule
func (c *APIClient) CreateAlertRule(req *AlertRuleCreateRequest) (*AlertRule, error) {
	respBody, statusCode, err := c.doRequest("POST", "/api-public/v1/alertRules", req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var rule AlertRule
	if err := json.Unmarshal(respBody, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// UpdateAlertRule updates an existing alert rule
func (c *APIClient) UpdateAlertRule(ruleID int, req *AlertRuleCreateRequest) (*AlertRule, error) {
	path := fmt.Sprintf("/api-public/v1/alertRules/%s", strconv.Itoa(ruleID))
	respBody, statusCode, err := c.doRequest("PUT", path, req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var rule AlertRule
	if err := json.Unmarshal(respBody, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// DeleteAlertRule deletes an alert rule
func (c *APIClient) DeleteAlertRule(ruleID int) error {
	path := fmt.Sprintf("/api-public/v1/alertRules/%s", strconv.Itoa(ruleID))
	respBody, statusCode, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 && statusCode != 204 {
		return fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	return nil
}
