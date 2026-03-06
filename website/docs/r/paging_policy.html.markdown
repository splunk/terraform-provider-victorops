---
layout: "victorops"
page_title: "VictorOps: victorops_paging_policy"
description: |-
  Manages a user's paging policy in VictorOps.
---

# victorops\_paging\_policy

Manages a [paging policy](https://portal.victorops.com/public/api-docs.html#!/PersonalPagingPolicies) for a VictorOps user. A paging policy defines how a user is notified when they are paged — the sequence of contact methods (phone, email, push notification) to attempt, and how long to wait between each step.

~> **Note:** The VictorOps API replaces the entire paging policy on every update. Destroying this resource clears the user's paging policy to an empty state.

## Example Usage

```hcl
resource "victorops_user" "alice" {
  first_name       = "Alice"
  last_name        = "Smith"
  user_name        = "asmith"
  email            = "asmith@example.com"
  is_admin         = false
  replacement_user = "admin"
}

resource "victorops_contact" "alice_phone" {
  user_name = victorops_user.alice.user_name
  type      = "phone"
  label     = "Mobile"
  value     = "+15551234567"
}

resource "victorops_paging_policy" "alice" {
  username = victorops_user.alice.user_name

  # Step 1: immediately send a push notification
  step {
    timeout = 0
    contact {
      type = "push"
    }
  }

  # Step 2: after 5 minutes, also call the phone
  step {
    timeout = 5
    contact {
      type       = "phone"
      contact_id = victorops_contact.alice_phone.internal_id
    }
  }

  # Step 3: after another 5 minutes, send an email
  step {
    timeout = 5
    contact {
      type       = "email"
      contact_id = victorops_user.alice.default_email_contact_id
    }
  }
}
```

## Argument Reference

* `username` - (Required, ForceNew) The VictorOps username whose paging policy this resource manages.
* `step` - (Required) An ordered list of paging steps. VictorOps tries each step in order, waiting `timeout` minutes before moving to the next.

### step

Each `step` block supports:

* `timeout` - (Required) Minutes to wait before escalating to the next step. Use `0` for the first step (immediate).
* `contact` - (Required) One or more contact methods to use in this step. At least one must be provided.

### contact

Each `contact` block within a `step` supports:

* `type` - (Required) The contact method type. Valid values: `"phone"`, `"email"`, `"push"`.
* `contact_id` - (Optional) The internal numeric ID of the contact method (from `victorops_contact.internal_id` or `victorops_user.default_email_contact_id`). Required for `phone` and `email` types; omit for `push`.

## Attributes Reference

* `id` - The username the paging policy belongs to (same as `username`).

## Import

An existing paging policy can be imported using the username:

```
terraform import victorops_paging_policy.alice asmith
```
