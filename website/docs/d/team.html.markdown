---
layout: "victorops"
page_title: "VictorOps: victorops_team (data source)"
description: |-
  Reads an existing VictorOps team.
---

# Data Source: victorops\_team

Use this data source to read information about an existing [VictorOps team](https://portal.victorops.com/public/api-docs.html#!/Teams/get_api_public_v1_team_team).

This is useful when you need to reference a team that already exists in VictorOps and was not created by Terraform — for example, to attach an escalation policy to it.

## Example Usage

```hcl
data "victorops_team" "platform" {
  slug = "team-abc123"
}

resource "victorops_escalation_policy" "pager" {
  name    = "Platform On-Call"
  team_id = data.victorops_team.platform.id
  step {
    timeout = 0
    entries = []
  }
}
```

## Argument Reference

* `slug` - (Required) The slug (unique identifier) of the team. This is the value shown in the VictorOps URL for the team, e.g. `team-abc123`.

## Attributes Reference

* `id` - The slug of the team (same as `slug`).
* `name` - The display name of the team.
* `member_count` - The number of members currently on the team.
* `is_default_team` - Whether this is the default team for the VictorOps organization.

## Import

Existing team resources can also be imported using the team slug:

```
terraform import victorops_team.example team-abc123
```
