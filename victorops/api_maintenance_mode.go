package victorops

import (
	"encoding/json"
	"fmt"
)

// MaintenanceModeState represents the maintenance mode state
type MaintenanceModeState struct {
	CompanyID       string                  `json:"companyId,omitempty"`
	ActiveInstances []ActiveMaintenanceMode `json:"activeInstances,omitempty"`
}

// ActiveMaintenanceMode represents an active maintenance mode instance
type ActiveMaintenanceMode struct {
	InstanceID string                  `json:"instanceId,omitempty"`
	StartedBy  string                  `json:"startedBy,omitempty"`
	StartedAt  int64                   `json:"startedAt,omitempty"`
	Targets    []MaintenanceModeTarget `json:"targets,omitempty"`
	IsGlobal   bool                    `json:"isGlobal,omitempty"`
	Purpose    string                  `json:"purpose,omitempty"`
}

// MaintenanceModeTarget represents a target for maintenance mode (spec lines 5508-5519)
type MaintenanceModeTarget struct {
	Type  string   `json:"type,omitempty"`  // "RoutingKeys"
	Names []string `json:"names,omitempty"` // routing key names
}

// StartMaintenanceModeRequest is the request body for starting maintenance mode
type StartMaintenanceModeRequest struct {
	Type    string   `json:"type"`
	Names   []string `json:"names"`
	Purpose string   `json:"purpose,omitempty"`
}

// GetMaintenanceModeState gets the current maintenance mode state
func (c *APIClient) GetMaintenanceModeState() (*MaintenanceModeState, error) {
	respBody, statusCode, err := c.doRequest("GET", "/api-public/v1/maintenancemode", nil)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var state MaintenanceModeState
	if err := json.Unmarshal(respBody, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// StartMaintenanceMode starts maintenance mode
func (c *APIClient) StartMaintenanceMode(req *StartMaintenanceModeRequest) (*MaintenanceModeState, error) {
	respBody, statusCode, err := c.doRequest("POST", "/api-public/v1/maintenancemode/start", req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var state MaintenanceModeState
	if err := json.Unmarshal(respBody, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// EndMaintenanceMode ends maintenance mode for the given instance ID
func (c *APIClient) EndMaintenanceMode(instanceID string) (*MaintenanceModeState, error) {
	path := fmt.Sprintf("/api-public/v1/maintenancemode/%s/end", instanceID)
	respBody, statusCode, err := c.doRequestText("PUT", path)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var state MaintenanceModeState
	if err := json.Unmarshal(respBody, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// FindMaintenanceModeByID finds a maintenance mode instance by ID
func (c *APIClient) FindMaintenanceModeByID(instanceID string) (*ActiveMaintenanceMode, error) {
	state, err := c.GetMaintenanceModeState()
	if err != nil {
		return nil, err
	}

	for _, instance := range state.ActiveInstances {
		if instance.InstanceID == instanceID {
			return &instance, nil
		}
	}

	return nil, nil
}
