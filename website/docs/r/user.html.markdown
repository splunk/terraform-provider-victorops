---
layout: "victorops"
page_title: "VictorOps: victorops_user"
description: |-
  Creates and manages a user in VictorOps.
---

# victorops\_user

A [user](https://portal.victorops.com/public/api-docs.html#/Users) is an individual within a VictorOps account.

Make sure the optional field 'replacement_user' is set to a default user_name to facilitate deleting users using TF. Alternatively, you can set the `VO_REPLACEMENT_USERNAME` env variable to the default username to replace all users when removed.

Note: We no longer allow creation of `admin` users through the Terraform (or the public API), the `is_admin` field value is ignored.
 

## Example Usage

```hcl
// Create a user within the victorops organization
resource "victorops_user" "user1" {
  first_name       = "Jane"
  last_name        = "Doe"
  user_name        = "JaneDoe"
  email            = "jdoe@victorops.com"
  is_admin         = true // depreacted and ignored. Cannot create admin users anymore.
  replacement_user = "myDefaultVOUser" // optional field
}

// Specify the replacement_user field with the default user_name to replace users when deleting them using TF

```

## Argument Reference

The following arguments are supported:

* `first_name` - (Required) The first name of the user.
* `last_name` - (Required) The last name of the user.
* `user_name` - (Required) The username for this user.
* `email` - (Required) The user's email address.
* `is_admin` - DEPRECATED - the field and its value will be ignored if specified.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the user.
* `default_email_contact_id` - The ID for the default email contact

## Import

Import is not currently supported
