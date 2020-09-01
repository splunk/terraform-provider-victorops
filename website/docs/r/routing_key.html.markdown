---
layout: "victorops"
page_title: "VictorOps: victorops_user"
description: |-
  Creates and manages a routing key in VictorOps
---

# victorops\_user

A routing key in VictorOps is used to route incoming alerts to specific escalation policies.

## Example Usage

```hcl
resource "victorops_routing_key" "infrastructure_high_severity" {
  name = "infrastructure-high-severity"
  targets = [victorops_escalation_policy.high_severity.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The first name of the user.
* `targets` - (Required) A list of escalation policy ids to route alerts to.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the routing key.

## Import

Import is not currently supported
