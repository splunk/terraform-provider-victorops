---
layout: "victorops"
page_title: "VictorOps: victorops_user"
description: |-
  Creates and manages a user in VictorOps.
---

# victorops\_team

A [team](https://portal.victorops.com/public/api-docs.html#/Teams) is a collection of users within a victorops account.They contain on-call rotations and escalation policies.

To add members to a team, use the [victorops_team_membership]('victorops_team_membership,html') resource.

## Example Usage

```hcl
# Create a team within victorops
resource "victorops_team" "infrastructure" {
  name = "Infrastructure"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of this team

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the team.

## Import

Import is not currently supported
