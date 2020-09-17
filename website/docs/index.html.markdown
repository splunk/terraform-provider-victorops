---
layout: "victorops"
page_title: "Provider: VictorOps"
description: |-
  Empower teams by routing alerts to the right people for fast collaboration and issue resolution.
---

# VictorOps Provider

[VictorOps](https://www.victorops.com) empowers teams by routing alerts to the right people for collaboration and fast issue resolution.

Using this VictorOps Terraform provider, teams can automate VictorOps setup associated with an application.
You can manage the following resources using this provider.
1. User
2. Team
3. User-Team assignment
4. Escalation Policy
5. Routing Key

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Intitialize the victorops provider

terraform {
	required_providers {
		victorops = {
			source = "splunk/victorops"
			version = "0.1.1"
		}
	}
}
  
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
