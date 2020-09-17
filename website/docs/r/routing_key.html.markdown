---
layout: "victorops"
page_title: "VictorOps: victorops_routing_key"
description: |-
  Creates and manages a routing key in VictorOps
---

# victorops\_routing\_key

A [routing key](https://portal.victorops.com/public/api-docs.html#!/Routing32Keys/get_api_public_v1_org_routing_keys) in VictorOps is used to route incoming alerts to specific escalation policies.

## Example Usage

```hcl
resource "victorops_routing_key" "infrastructure_high_severity" {
  name = "infrastructure-high-severity"
  targets = [victorops_escalation_policy.high_severity.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Routing key name.
* `targets` - (Required) A list of escalation policy ids to route alerts to.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the routing key.

## Import

Import is not currently supported
