---
layout: "victorops"
page_title: "VictorOps: victorops_contact"
description: |-
  Manage a user's contact methods in VictorOps.
---

# victorops\_contact

A contact represents emails and phone numbers that can be used to contact a user.

NOTE: When adding a contact phone through terraform, a user must manually verify the phone number in the web ui before
it can be used in a paging policy.

## Example Usage

```hcl
resource "victorops_contact" "jenny_cell" {
  user_name    = victorops_user.jenny.user_name
  type         = "phone"
  value        = "+13038674309"
  label        = "Jenny's Cell Phone"
}

resource "victorops_contact" "jenny_email" {
  user_name    = victorops_user.jenny.user_name
  type         = "email"
  value        = "jenny@victorops.com"
  label        = "Jenny's Work Email"
}
```

## Argument Reference

The following arguments are supported:

* `user_name` - (Required) The username of the user this phone number belongs to.
* `value` - (Required) A phone number or email address that will be used for this contact.
* `type`  - (Required) Either "phone" or "email" depending on the type of contact you are creating.
* `label` - (Required) What label to give this contact. Example: work

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the contact.

## Import

Import is not currently supported
