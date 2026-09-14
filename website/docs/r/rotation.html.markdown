---
layout: "victorops"
page_title: "VictorOps: victorops_rotation"
description: |-
  Creates and manages an on-call rotation in VictorOps.
---

# victorops\_rotation

Creates and manages a rotation group, its shifts, and the ordered members of each shift. The exported rotation slug can be used by `victorops_escalation_policy`.

## Example Usage

```hcl
resource "victorops_rotation" "primary" {
  team_id = victorops_team.infrastructure.id
  name    = "Primary"

  shift {
    name     = "Primary 24x7"
    timezone = "America/Toronto"
    start    = 1767225600000
    duration = 7
    type     = "std"
    members  = [
      victorops_user.alice.user_name,
      victorops_user.bob.user_name,
    ]

    mask {
      days = ["su", "m", "t", "w", "th", "f", "sa"]

      time {
        start_hour   = 0
        start_minute = 0
        end_hour     = 23
        end_minute   = 59
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) The slug of the team that owns the rotation. Changing it creates a new rotation.
* `name` - (Required) The rotation group label.
* `shift` - (Optional) One or more ordered shift blocks.

Each `shift` block supports:

* `name` - (Required) The shift label.
* `timezone` - (Required) An IANA timezone such as `America/Toronto`.
* `start` - (Required) The shift start time in epoch milliseconds.
* `duration` - (Required) The duration in days, from 1 through 90.
* `type` - (Required) `std`, `fts`, `cstm`, or `pho`.
* `members` - (Optional) An ordered list of VictorOps usernames.
* `mask` - (Required) The primary schedule mask.
* `mask2` - (Optional) A second mask used by custom and follow-the-sun shifts.
* `mask3` - (Optional) A third mask used by custom and follow-the-sun shifts.

Each mask supports a `days` set containing `su`, `m`, `t`, `w`, `th`, `f`, and/or `sa`, plus one or more `time` blocks. A `time` block has `start_hour`, `start_minute`, `end_hour`, and `end_minute`.

## Attributes Reference

The following attributes are exported:

* `id` - The numeric rotation group ID.
* `slug` - The rotation group slug used by escalation policies.
* `shift.id` - The numeric shift ID.

## Import

Import a rotation using its team slug and numeric rotation group ID:

```sh
terraform import victorops_rotation.primary team-AbCdEf0123456789/12345
```
