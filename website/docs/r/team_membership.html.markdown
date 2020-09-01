---
layout: "victorops"
page_title: "VictorOps: victorops_contact_phone"
description: |-
  Manages a user's team association within victorops
---

# victorops\_team\_membership

Associates a user with a specific team in VictorOps

## Example Usage

```hcl
resource "victorops_team_membership" "john_infra" {
  team_id          = victorops_team.infrastructure.id
  user_name        = victorops_user.john.user_name
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) The id of the team we are adding this user to.
* `user_name` - (Required) The username of the victorops user to add to the team.

## Attributes Reference

There are no additional attributes for this resource.

## Import

Import is not currently supported
