package victorops

import (
	"encoding/json"
	"fmt"
)

// TeamUpdateRequest is the request body for updating a team
type TeamUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// TeamResponse is the API response for team operations
type TeamResponse struct {
	Name          string `json:"name,omitempty"`
	Slug          string `json:"slug,omitempty"`
	MemberCount   int    `json:"memberCount,omitempty"`
	Version       int    `json:"version,omitempty"`
	IsDefaultTeam bool   `json:"isDefaultTeam,omitempty"`
	Description   string `json:"description,omitempty"`
}

// UpdateTeam updates a team's name and/or description
// This bypasses the buggy go-victorops library which incorrectly uses team.Name for URL path
func (c *APIClient) UpdateTeam(teamSlug string, req *TeamUpdateRequest) (*TeamResponse, error) {
	path := fmt.Sprintf("/api-public/v1/team/%s", teamSlug)
	respBody, statusCode, err := c.doRequest("PUT", path, req)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", statusCode, string(respBody))
	}

	var team TeamResponse
	if err := json.Unmarshal(respBody, &team); err != nil {
		return nil, err
	}

	return &team, nil
}
