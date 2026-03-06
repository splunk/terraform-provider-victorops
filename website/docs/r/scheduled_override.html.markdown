---
layout: "victorops"
page_title: "VictorOps: victorops_scheduled_override"
description: |-
  Creates a temporary on-call override in VictorOps.
---

# victorops\_scheduled\_override

Creates a [scheduled on-call override](https://portal.victorops.com/public/api-docs.html#!/OnCallOverrides) in VictorOps. This allows you to temporarily replace the on-call user for a specific time window — for example, covering a vacation or planned maintenance.

When this resource is destroyed, the override is deleted and the original on-call rotation resumes.

~> **Note:** Because the VictorOps API does not support updating overrides, all fields are `ForceNew`. Any change to the override window requires destroying and recreating the resource.

## Example Usage

```hcl
# Cover a user's on-call shift for a long weekend
resource "victorops_scheduled_override" "holiday_cover" {
  username    = "jsmith"
  start       = "2024-08-31T00:00:00Z"
  end         = "2024-09-03T00:00:00Z"
  policy_slug = victorops_escalation_policy.high_severity.id
}
```

## Argument Reference

* `username` - (Required, ForceNew) The username of the user taking over on-call duty.
* `start` - (Required, ForceNew) The start of the override window in RFC3339 format (e.g. `"2024-08-31T08:00:00Z"`).
* `end` - (Required, ForceNew) The end of the override window in RFC3339 format.
* `policy_slug` - (Optional, ForceNew) The slug of the escalation policy to restrict the override to. If omitted, the override applies to all policies the original user is on-call for.

## Attributes Reference

* `id` - The public ID assigned to the override by VictorOps.
* `original_on_call_user` - The username of the original on-call user who was replaced by this override.

## Import

Scheduled overrides can be imported using the public ID:

```
terraform import victorops_scheduled_override.example abc-xyz-123
```
