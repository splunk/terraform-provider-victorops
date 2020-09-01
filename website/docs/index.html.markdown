---
layout: "victorops"
page_title: "Provider: VictorOps"
description: |-
  VictorOps VictorOps incident management software gives DevOps observability, collaboration, & real-time alerting, to build, deploy, & operate software.
---

# VictorOps Provider

[VictorOps](https://www.victorops.com) is an alarm aggregation and dispatching service for system administrators and support teams. It collects alerts from your monitoring tools, gives you an overall view of all of your monitoring alarms, and alerts an on duty engineer if there’s a problem.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Intitialize the victorops provider
provider "victorops" {
  api_id  = var.api_id
  api_key = var.api_key
}

# Create a team within victorops
resource "victorops_team" "infrastructure" {
  name = "Infrastructure"
}

# Create a user within the victorops organization
resource "victorops_user" "user1" {
  first_name       = "Jane"
  last_name        = "Doe"
  user_name        = "JaneDoe"
  email            = "jdoe@victorops.com"
  is_admin         = true
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) An API id tied to an admin user
* `api_key` - (Required) An API key tied to an admin user
