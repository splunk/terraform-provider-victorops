---
layout: "victorops"
page_title: "VictorOps: victorops_escalation_policy"
description: |-
  Creates and manages a user in VictorOps.
---

# victorops\_escalation\_policy

[Team Escalation Policies](https://portal.victorops.com/public/api-docs.html#/Escalation32Policies) set who is actually on-call for a given team and are the link to utilize any rotations that have been created.

Note: You need to fetch an existing Rotation Group Slug through the VO public API - [GET-Rotations](https://portal.victorops.com/public/api-docs.html#!/Rotations/get_api_public_v1_teams_team_rotations) for creating an escalation policy resource from Terraform

## Example Usage

```hcl
resource "victorops_escalation_policy" "vikings_high_severity" {
  name    = "High Severity"
  team_id = victorops_team.team_vikings.id
  step {
    timeout = 0
    entries = [
      {
        type = "rotationGroup"
        slug = "rtg-wvvhXshpvaRdn7jM"
      }
    ]
  }
  step {
    timeout = 10
    entries = [
      {
        type = "rotationGroup"
        slug = "rtg-hfy3fUytq7otMNbf"
      }
    ]
  }
}

resource "victorops_escalation_policy" "vikings_low_severity" {
  name    = "Low Severity"
  team_id = victorops_team.team_vikings.id
  step {
    timeout = 0
    entries = [
      {
        type = "rotationGroup"
        slug = "rtg-wvvhXshpvaRdn7jM"
      }
    ]
  }
  step {
    timeout = 300
    entries = [
      {
        type = "rotationGroup"
        slug = "rtg-hfy3fUytq7otMNbf"
      }
    ]
  }
  step {
    timeout = 300
    entries = [
      {
        type = "targetPolicy"
        slug = victorops_escalation_policy.vikings_high_severity.id
      }
    ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of this escalation policy
* `team_id` - (Required) The team_id of the team for which you want to create this escalation policy
* `ignore_custom_paging_policies` - (Optional) `true`/`false`
* `step` - (Required) - The escalation policy step defined in the following structure

```hcl
	step {
		timeout = [time-out duration in seconds]
		entries = [
			{
				type = [ rotationalGroup | targetPolicy ]
				slug = [ rotatioGroup slug | next escalation policy ID ]
			},
		]
	}
```

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the escalation policy.

## Import

Import is not currently supported
