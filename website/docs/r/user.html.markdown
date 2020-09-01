---
layout: "victorops"
page_title: "VictorOps: victorops_user"
description: |-
  Creates and manages a user in VictorOps.
---

# victorops\_user

A [user](https://portal.victorops.com/public/api-docs.html#/Users) is an individual within a VictorOps account.

## Example Usage

```hcl
// Create a user within the victorops organization
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

* `frist_name` - (Required) The first name of the user.
* `last_name` - (Required) The last name of the user.
* `user_name` - (Required) The username for this user.
* `email` - (Required) The user's email address.
* `is_admin` - (Optional, Default: false) If this user is an account admin.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the user.
* `default_email_contact_id` - The ID for the default email contact

## Import

Import is not currently supported
