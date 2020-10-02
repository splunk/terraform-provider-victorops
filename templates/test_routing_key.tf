

resource "victorops_team" "test_team_1" {
  name = "{{.TeamName}}"
}

resource "victorops_escalation_policy" "test_policy_2" {
  name    = "{{.PolicyName}}"
  team_id = victorops_team.test_team_1.id
  step {
    timeout = 60
    entries = []
  }
}

// Create routing keys to push alerts to our escalation policies
resource "victorops_routing_key" "test_key" {
  name = "{{.RoutingKeyName}}"
  targets = [victorops_escalation_policy.test_policy_2.id]
}
