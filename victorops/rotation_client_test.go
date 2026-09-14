package victorops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestRotationClientCRUD(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-VO-Api-Id") != "api-id" || r.Header.Get("X-VO-Api-Key") != "api-key" {
			t.Errorf("rotation request did not include VictorOps authentication headers")
		}
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "POST /api-public/v1/teams/team-test/rotations":
			writeRotationJSON(t, w, map[string]interface{}{"id": 42, "label": "Primary"})
		case "GET /api-public/v1/teams/team-test/rotations/42":
			writeRotationJSON(t, w, testRotationGroupResponse())
		case "GET /api-public/v1/teams/team-test/rotations":
			writeRotationJSON(t, w, map[string]interface{}{
				"rotationGroups": []interface{}{map[string]interface{}{
					"groupId": "42", "label": "Primary", "slug": "rtg-1234567890abcdef",
				}},
			})
		case "PUT /api-public/v1/teams/team-test/rotations/42":
			assertRotationJSONField(t, r, "label", "Renamed")
			writeRotationJSON(t, w, testRotationGroupResponse())
		case "DELETE /api-public/v1/teams/team-test/rotations/42":
			w.WriteHeader(http.StatusOK)
		case "POST /api-public/v1/teams/team-test/rotations/42":
			writeRotationJSON(t, w, map[string]interface{}{"rot_id": 7, "group_id": 42})
		case "PUT /api-public/v1/teams/team-test/rotations/42/7":
			w.WriteHeader(http.StatusOK)
		case "DELETE /api-public/v1/teams/team-test/rotations/42/7":
			w.WriteHeader(http.StatusOK)
		case "POST /api-public/v1/teams/team-test/rotations/42/7/members":
			assertRotationJSONField(t, r, "username", "alice")
			w.WriteHeader(http.StatusOK)
		case "PUT /api-public/v1/teams/team-test/rotations/42/7/members":
			assertRotationJSONField(t, r, "position", float64(0))
			w.WriteHeader(http.StatusOK)
		case "DELETE /api-public/v1/teams/team-test/rotations/42/7/members":
			assertRotationJSONField(t, r, "username", "alice")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newRotationClient("api-id", "api-key", server.URL)
	client.httpClient = server.Client()
	client.minInterval = 0
	payload := testRotationShiftPayload()

	group, status, _, err := client.createGroup("team-test", rotationGroupPayload{Label: "Primary"})
	assertRotationResult(t, status, err)
	if group.ID != 42 {
		t.Fatalf("expected group ID 42, got %d", group.ID)
	}
	group, status, _, err = client.getGroup("team-test", 42)
	assertRotationResult(t, status, err)
	if len(group.Shifts) != 1 || group.Shifts[0].RotationID != 7 {
		t.Fatalf("unexpected group response: %#v", group)
	}
	groups, status, _, err := client.listGroups("team-test")
	assertRotationResult(t, status, err)
	if len(groups.RotationGroups) != 1 || groups.RotationGroups[0].Slug != "rtg-1234567890abcdef" {
		t.Fatalf("unexpected group list response: %#v", groups)
	}
	status, _, err = client.updateGroup("team-test", 42, "Renamed")
	assertRotationResult(t, status, err)
	shift, status, _, err := client.createShift("team-test", 42, payload)
	assertRotationResult(t, status, err)
	if shift.RotationID != 7 {
		t.Fatalf("expected shift ID 7, got %d", shift.RotationID)
	}
	status, _, err = client.updateShift("team-test", 42, 7, payload)
	assertRotationResult(t, status, err)
	status, _, err = client.addMember("team-test", 42, 7, "alice")
	assertRotationResult(t, status, err)
	status, _, err = client.positionMember("team-test", 42, 7, "alice", 0)
	assertRotationResult(t, status, err)
	status, _, err = client.removeMember("team-test", 42, 7, "alice")
	assertRotationResult(t, status, err)
	status, _, err = client.deleteShift("team-test", 42, 7)
	assertRotationResult(t, status, err)
	status, _, err = client.deleteGroup("team-test", 42)
	assertRotationResult(t, status, err)

	if len(requests) != 11 {
		t.Fatalf("expected 11 API requests, got %d: %v", len(requests), requests)
	}
}

func TestRotationExpandAndFlatten(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRotation().Schema, testRotationResourceConfig())

	shifts, err := expandRotationShifts(resourceData.Get("shift"))
	if err != nil {
		t.Fatal(err)
	}
	if len(shifts) != 1 || shifts[0].Payload.Start != 1767225600000 || !reflect.DeepEqual(shifts[0].Members, []string{"alice", "bob"}) {
		t.Fatalf("unexpected expanded shifts: %#v", shifts)
	}
	if !shifts[0].Payload.Mask.Day.Monday || !shifts[0].Payload.Mask.Day.Tuesday {
		t.Fatalf("expected Monday and Tuesday in expanded mask: %#v", shifts[0].Payload.Mask)
	}

	response := rotationShiftResponse{rotationShiftPayload: shifts[0].Payload, RotationID: 7, Usernames: shifts[0].Members}
	flattened := flattenRotationShifts([]rotationShiftResponse{response})
	if err := resourceData.Set("shift", flattened); err != nil {
		t.Fatal(err)
	}
	if got := resourceData.Get("shift.0.id"); got != 7 {
		t.Fatalf("expected flattened shift ID 7, got %#v", got)
	}
}

func TestResourceRotationCreateReadDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST /api-public/v1/teams/team-test/rotations":
			var payload rotationGroupPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.Label != "Primary" || len(payload.Shifts) != 1 || !reflect.DeepEqual(payload.Shifts[0].Usernames, []string{"alice", "bob"}) {
				t.Errorf("unexpected create payload: %#v", payload)
			}
			writeRotationJSON(t, w, testRotationGroupResponse())
		case "GET /api-public/v1/teams/team-test/rotations/42":
			if deleted {
				http.NotFound(w, r)
				return
			}
			writeRotationJSON(t, w, testRotationGroupResponse())
		case "GET /api-public/v1/teams/team-test/rotations":
			writeRotationJSON(t, w, map[string]interface{}{
				"rotationGroups": []interface{}{map[string]interface{}{
					"groupId": "42", "label": "Primary", "slug": "rtg-1234567890abcdef",
				}},
			})
		case "DELETE /api-public/v1/teams/team-test/rotations/42":
			deleted = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newRotationClient("api-id", "api-key", server.URL)
	client.httpClient = server.Client()
	client.minInterval = 0
	d := schema.TestResourceDataRaw(t, resourceRotation().Schema, testRotationResourceConfig())
	config := Config{APIId: "api-id", APIKey: "api-key", BaseURL: server.URL, RotationClient: client}

	if err := resourceRotationCreate(d, config); err != nil {
		t.Fatal(err)
	}
	if d.Id() != "42" || d.Get("slug").(string) != "rtg-1234567890abcdef" {
		t.Fatalf("unexpected rotation state: id=%q slug=%q", d.Id(), d.Get("slug"))
	}
	if err := resourceRotationDelete(d, config); err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("expected rotation to be deleted")
	}
}

func TestReconcileRotationShifts(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "PUT /api-public/v1/teams/team-test/rotations/42/7":
			w.WriteHeader(http.StatusOK)
		case "POST /api-public/v1/teams/team-test/rotations/42":
			writeRotationJSON(t, w, map[string]interface{}{"rot_id": 9, "group_id": 42})
		case "POST /api-public/v1/teams/team-test/rotations/42/7/members",
			"PUT /api-public/v1/teams/team-test/rotations/42/7/members",
			"POST /api-public/v1/teams/team-test/rotations/42/9/members",
			"PUT /api-public/v1/teams/team-test/rotations/42/9/members",
			"DELETE /api-public/v1/teams/team-test/rotations/42/8":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newRotationClient("api-id", "api-key", server.URL)
	client.httpClient = server.Client()
	client.minInterval = 0
	payload := testRotationShiftPayload()
	current := []rotationShiftResponse{
		{rotationShiftPayload: payload, RotationID: 7, Usernames: []string{"alice"}},
		{rotationShiftPayload: payload, RotationID: 8, Usernames: []string{"obsolete"}},
	}
	desired := []rotationShiftConfig{
		{ID: 7, Payload: payload, Members: []string{"bob", "alice"}},
		{Payload: payload, Members: []string{"carol"}},
	}

	if err := reconcileRotationShifts(client, "team-test", 42, current, desired); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 8 {
		t.Fatalf("expected 8 reconciliation requests, got %d: %v", len(requests), requests)
	}
}

func testRotationResourceConfig() map[string]interface{} {
	return map[string]interface{}{
		"team_id": "team-test",
		"name":    "Primary",
		"shift": []interface{}{map[string]interface{}{
			"name": "Primary 24x7", "timezone": "America/Toronto", "start": 1767225600000,
			"duration": 7, "type": "std", "members": []interface{}{"alice", "bob"},
			"mask": []interface{}{map[string]interface{}{
				"days": []interface{}{"m", "t"},
				"time": []interface{}{map[string]interface{}{
					"start_hour": 9, "start_minute": 30, "end_hour": 17, "end_minute": 0,
				}},
			}},
		}},
	}
}

func TestRotationImport(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceRotation().Schema, map[string]interface{}{})
	d.SetId("team-test/42")
	resources, err := resourceRotationImport(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 1 || d.Id() != "42" || d.Get("team_id").(string) != "team-test" {
		t.Fatalf("unexpected imported rotation state: id=%q team=%q", d.Id(), d.Get("team_id"))
	}
}

func testRotationShiftPayload() rotationShiftPayload {
	return rotationShiftPayload{
		Label: "Primary 24x7", Timezone: "America/Toronto", Start: 1767225600000,
		Duration: 7, ShiftType: "std",
		Mask: rotationMask{
			Day:  rotationDays{Monday: true},
			Time: []rotationTimeRange{{Start: rotationTime{Hour: 0}, End: rotationTime{Hour: 23, Minute: 59}}},
		},
	}
}

func testRotationGroupResponse() map[string]interface{} {
	return map[string]interface{}{
		"id": 42, "label": "Primary",
		"shifts": []interface{}{map[string]interface{}{
			"rot_id": 7, "group_id": 42, "label": "Primary 24x7", "timezone": "America/Toronto",
			"start": 1767225600000, "duration": 7, "shifttype": "std", "usernames": []string{"alice"},
			"mask": map[string]interface{}{
				"day": map[string]bool{"m": true},
				"time": []interface{}{map[string]interface{}{
					"start": map[string]int{"hour": 0, "minute": 0},
					"end":   map[string]int{"hour": 23, "minute": 59},
				}},
			},
		}},
	}
}

func writeRotationJSON(t *testing.T, w http.ResponseWriter, value interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("failed to encode test response: %v", err)
	}
}

func assertRotationJSONField(t *testing.T, r *http.Request, field string, expected interface{}) {
	t.Helper()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("invalid JSON request %q: %v", string(body), err)
	}
	if !reflect.DeepEqual(payload[field], expected) {
		t.Errorf("expected %s=%#v, got %#v", field, expected, payload[field])
	}
}

func assertRotationResult(t *testing.T, status int, err error) {
	t.Helper()
	if err != nil || status != http.StatusOK {
		t.Fatalf("unexpected rotation API result: status=%d error=%v", status, err)
	}
}

func TestRotationClientRejectsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not-json")
	}))
	defer server.Close()
	client := newRotationClient("api-id", "api-key", server.URL)
	client.httpClient = server.Client()
	client.minInterval = 0 * time.Second
	_, _, _, err := client.getGroup("team-test", 42)
	if err == nil {
		t.Fatal("expected invalid JSON response to return an error")
	}
}
