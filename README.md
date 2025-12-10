Terraform `VictorOps/Splunk OnCall` Provider
=========================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

- A VictorOps/Splunk OnCall account with API access (API key and ID).
- [Terraform](https://www.terraform.io/downloads.html) 0.13.x or higher
- [Go](https://golang.org/doc/install) 1.21+ (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/splunk/terraform-provider-victorops`

```sh
$ git clone git@github.com:splunk/terraform-provider-victorops.git $GOPATH/src/github.com/splunk/terraform-provider-victorops
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/splunk/terraform-provider-victorops
$ go build -o terraform-provider-victorops
```

Features
------------

Using this VictorOps/Splunk OnCall Terraform provider, you can manage the following resources:

### Resources

| Resource | Description |
|----------|-------------|
| `victorops_user` | Manage users |
| `victorops_team` | Manage teams |
| `victorops_team_membership` | Manage user-team assignments |
| `victorops_escalation_policy` | Manage escalation policies |
| `victorops_routing_key` | Manage routing keys |
| `victorops_user_contact_email` | Manage user email contact methods |
| `victorops_user_contact_phone` | Manage user phone contact methods |
| `victorops_alert_rule` | Manage alert routing rules |
| `victorops_maintenance_mode` | Manage maintenance windows |
| `victorops_scheduled_override` | Manage scheduled on-call overrides |
| `victorops_user_paging_policy` | Manage user paging policies |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `victorops_users` | List users (with optional email filter) |
| `victorops_routing_keys` | List routing keys |
| `victorops_team_admins` | List team administrators |
| `victorops_user_devices` | List user contact devices (read-only) |
| `victorops_rotations` | List team rotations |
| `victorops_team_oncall_schedule` | Get team on-call schedule |

Usage
------------

### Provider Configuration

```hcl
terraform {
  required_providers {
    victorops = {
      source  = "splunk/victorops"
      version = "~> 0.2.0"
    }
  }
}

provider "victorops" {
  api_id  = var.victorops_api_id   # Or set VO_API_ID env var
  api_key = var.victorops_api_key  # Or set VO_API_KEY env var
}
```

### Example: Managing Users and Teams

```hcl
# Create a user
resource "victorops_user" "jdane" {
  first_name       = "John"
  last_name        = "Dane"
  user_name        = "jdane"
  email            = "jdane@example.com"
  is_admin         = false           # deprecated - We no longer support creating admin users through TF/public APIs. The value in this field is ignored.
  replacement_user = "default_user"  # optional - Specify the default username to replace all users when deleting users using TF
}

# Create a team
resource "victorops_team" "platform" {
  name = "Platform-Team"
}

# Add user to team
resource "victorops_team_membership" "jdane_platform" {
  team_id   = victorops_team.platform.id
  user_name = victorops_user.jdane.user_name
}
```

### Example: Contact Methods

```hcl
# Add email contact method
resource "victorops_user_contact_email" "jdane_work" {
  username = victorops_user.jdane.user_name
  email    = "jdane-alerts@example.com"
  label    = "Work Email"
}

# Add phone contact method
resource "victorops_user_contact_phone" "jdane_mobile" {
  username = victorops_user.jdane.user_name
  phone    = "+1-555-123-4567"
  label    = "Mobile"
}
```

### Example: Escalation Policy and Routing

```hcl
# Create escalation policy
resource "victorops_escalation_policy" "high_severity" {
  name    = "High Severity"
  team_id = victorops_team.platform.id
  step {
    timeout = 60
    entries = [
      {
        type = "rotationGroup"
        slug = "rtg-wvvhXshpvaRdn7jM"
      }
    ]
  }
}

# Create routing key
resource "victorops_routing_key" "platform_alerts" {
  name    = "platform-alerts"
  targets = [victorops_escalation_policy.high_severity.id]
}
```

### Example: Alert Rules

```hcl
# Create alert routing rule
resource "victorops_alert_rule" "database_alerts" {
  alert_field       = "host_name"
  alert_value_match = "db-*"
  match_type        = "WILDCARD"
  routing_key       = victorops_routing_key.platform_alerts.name
  stop_flag         = false
  notes             = "Route database alerts to platform team"
}
```

### Example: Maintenance Mode

```hcl
# Start maintenance mode for specific routing keys
resource "victorops_maintenance_mode" "deploy" {
  routing_keys = [victorops_routing_key.platform_alerts.name]
  purpose      = "Scheduled deployment"
}

# Global maintenance mode (all routing keys)
resource "victorops_maintenance_mode" "global" {
  purpose = "Infrastructure maintenance"
}
```

### Example: Scheduled Override

```hcl
# Schedule an on-call override
resource "victorops_scheduled_override" "vacation" {
  username = "jdane"
  timezone = "America/New_York"
  start    = "2024-12-20T09:00:00Z"
  end      = "2024-12-27T09:00:00Z"
}
```

### Example: Data Sources

```hcl
# List all users
data "victorops_users" "all" {}

# List users by email
data "victorops_users" "by_email" {
  email = "jdane@example.com"
}

# List routing keys
data "victorops_routing_keys" "all" {}

# Get team on-call schedule
data "victorops_team_oncall_schedule" "platform" {
  team_id      = victorops_team.platform.id
  days_forward = 14
}

# List team rotations
data "victorops_rotations" "platform" {
  team_id = victorops_team.platform.id
}

# List team admins
data "victorops_team_admins" "platform" {
  team_id = victorops_team.platform.id
}

# List user devices (read-only)
data "victorops_user_devices" "jdane" {
  username = "jdane"
}
```

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.21+ is *required*).

To compile the provider:

```sh
$ go build -o terraform-provider-victorops
```

To run unit tests:

```sh
$ go test -short ./victorops/...
```

To run acceptance tests, set the following environment variables:

| Variable | Description |
|----------|-------------|
| `VO_API_ID` | VictorOps API ID |
| `VO_API_KEY` | VictorOps API Key |
| `VO_BASE_URL` | API base URL (default: `https://api.victorops.com`) |
| `VO_REPLACEMENT_USERNAME` | Default username to replace deleted users |
| `VO_OVERRIDE_USERNAME` | (Optional) Username in an escalation policy for override tests |

```sh
$ export VO_API_ID="your-api-id"
$ export VO_API_KEY="your-api-key"
$ export VO_REPLACEMENT_USERNAME="existing-user"
$ TF_ACC=1 go test -v ./victorops/... -timeout 120m
```

API Rate Limiting
-----------------

The VictorOps API has a rate limit of approximately 2 requests per second. This provider implements automatic retry with exponential backoff for rate-limited requests (HTTP 429) and transient server errors (HTTP 500, 502, 503, 504).

License
-------

See [LICENSE](LICENSE) file.
