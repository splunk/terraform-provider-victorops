resource "victorops_team" "test_team" {
  name = "{{.TeamName}}"
}

resource "victorops_escalation_policy" "test_policy" {
  name    = "{{.PolicyName}}"
  team_id = victorops_team.test_team.id
  step {
    timeout = 60
    entries = [
      {
        type = "rotationGroup"
        slug = "{{.RotationSlug}}"
      }
    ]
  }
}
