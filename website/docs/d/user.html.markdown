---
layout: "victorops"
page_title: "VictorOps: victorops_user (data source)"
description: |-
  Reads an existing VictorOps user.
---

# Data Source: victorops\_user

Use this data source to read information about an existing [VictorOps user](https://portal.victorops.com/public/api-docs.html#!/Users/get_api_public_v1_user_user).

This is useful when you need to reference a user that already exists in VictorOps and was not created by Terraform.

## Example Usage

```hcl
data "victorops_user" "ops_lead" {
  user_name = "jsmith"
}

output "ops_lead_email" {
  value = data.victorops_user.ops_lead.email
}
```

## Argument Reference

* `user_name` - (Required) The username of the VictorOps user to look up.

## Attributes Reference

* `id` - The username of the user (same as `user_name`).
* `first_name` - The user's first name.
* `last_name` - The user's last name.
* `email` - The user's primary email address.
* `default_email_contact_id` - The internal numeric ID of the user's default email contact method. Useful when building paging policies.
