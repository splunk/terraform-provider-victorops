---
layout: "victorops"
page_title: "VictorOps: victorops_rotation"
sidebar_current: "docs-victorops-resource-rotation"
description: |-
  Manages a VictorOps rotation group and its shifts.
---

# victorops_rotation

Manages a VictorOps rotation group (and its shifts) under a team.

The rotation group `slug` (e.g. `rtg-abc123`) is exposed as a computed attribute
and can be referenced directly in `victorops_escalation_policy` step entries.

> **Note:** There is no update endpoint in the VictorOps public API for rotation groups.
> All fields are `ForceNew` — any change destroys and recreates the rotation group.

## Example Usage

```hcl
resource "victorops_rotation" "primary" {
  team_slug = "team-abc123"
  label     = "Primary On-Call"

  shift {
    label     = "Weekly Shift"
    timezone  = "America/New_York"
    start     = 1704067200000  # epoch milliseconds
    duration  = 7              # 7-day rotation
    shifttype = "std"
    usernames = ["alice", "bob"]

    mask {
      monday    = true
      tuesday   = true
      wednesday = true
      thursday  = true
      friday    = true

      time_range {
        start_hour   = 9
        start_minute = 0
        end_hour     = 17
        end_minute   = 0
      }
    }
  }
}

# Use the rotation slug in an escalation policy
resource "victorops_escalation_policy" "oncall" {
  name    = "Primary Policy"
  team_id = "team-abc123"

  step {
    timeout = 0
    entries = [
      {
        type = "rotationGroup"
        slug = victorops_rotation.primary.slug
      }
    ]
  }
}
```

## Argument Reference

* `team_slug` - (Required, ForceNew) The slug of the team that owns this rotation group.
* `label` - (Required, ForceNew) The display name of the rotation group.
* `shift` - (Optional, ForceNew) One or more shift blocks. See [Shift](#shift) below.

### Shift

* `label` - (Required) The name of the shift.
* `timezone` - (Required) IANA timezone name, e.g. `"America/Denver"` or `"Europe/London"`.
* `start` - (Required) Shift start time as Unix epoch **milliseconds**.
* `duration` - (Required) Rotation period in days (1–90).
* `shifttype` - (Required) One of: `std` (Standard), `fts` (Follow The Sun), `cstm` (Custom), `pho` (Period Hand Off).
* `usernames` - (Optional) Ordered list of usernames in the rotation.
* `mask` - (Required) Primary active-period mask. See [Mask](#mask) below.
* `mask2` - (Optional) Secondary mask, used with `fts` and `cstm` shift types.
* `mask3` - (Optional) Tertiary mask, used with `cstm` shift type.

### Mask

Defines which days and times the rotation is active.

* `sunday` - (Optional) Default `false`.
* `monday` - (Optional) Default `true`.
* `tuesday` - (Optional) Default `true`.
* `wednesday` - (Optional) Default `true`.
* `thursday` - (Optional) Default `true`.
* `friday` - (Optional) Default `true`.
* `saturday` - (Optional) Default `false`.
* `time_range` - (Optional) One or more time range blocks. If omitted, the shift is 24×7 on the active days.

#### time_range

* `start_hour` - (Required) Hour of day the shift starts (0–23).
* `start_minute` - (Required) Minute the shift starts (0–59).
* `end_hour` - (Required) Hour of day the shift ends (0–23).
* `end_minute` - (Required) Minute the shift ends (0–59).

## Attributes Reference

In addition to all arguments above:

* `id` - The resource ID in the format `team_slug/groupId`.
* `slug` - The unique rotation group slug (e.g. `rtg-abc123`). Use this value in escalation policy step entries.

## Import

Rotation groups can be imported using `team_slug/groupId`:

```
terraform import victorops_rotation.primary team-abc123/42
```

The numeric `groupId` can be found from the API:

```bash
curl -H "X-VO-Api-Id: $VO_API_ID" -H "X-VO-Api-Key: $VO_API_KEY" \
     https://api.victorops.com/api-public/v1/teams/team-abc123/rotations
```
