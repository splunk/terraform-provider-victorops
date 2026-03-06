---
layout: "victorops"
page_title: "VictorOps: victorops_escalation_policy (data source)"
description: |-
  Reads an existing VictorOps escalation policy.
---

# Data Source: victorops\_escalation\_policy

Use this data source to read information about an existing [VictorOps escalation policy](https://portal.victorops.com/public/api-docs.html#!/EscalationPolicies).

This is useful when you need to reference a policy that was created outside of Terraform — for example, to add it as a routing key target.

## Example Usage

```hcl
data "victorops_escalation_policy" "high_severity" {
  slug = "pol-abc123"
}

resource "victorops_routing_key" "alerts" {
  name    = "prod-alerts"
  targets = [data.victorops_escalation_policy.high_severity.id]
}
```

## Argument Reference

* `slug` - (Required) The slug (unique identifier) of the escalation policy, e.g. `pol-abc123`.

## Attributes Reference

* `id` - The slug of the policy (same as `slug`).
* `name` - The display name of the escalation policy.
* `team_slug` - The slug of the team this policy belongs to.

## Import

Existing escalation policy resources can be imported using the policy slug:

```
terraform import victorops_escalation_policy.example pol-abc123
```
