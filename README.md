Terraform `VictorOps` Provider
=========================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

- A VictorOps account that you want to manage alongwith API key and token.
- [Terraform](https://www.terraform.io/downloads.html) 0.10.x
- [Go](https://golang.org/doc/install) 1.14 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/splunk/terraform-provider-victorops`

```sh
$ git clone git@github.com:splunk/terraform-provider-victorops.git $GOPATH/src/github.com/splunk/terraform-provider-victorops
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/splunk/terraform-provider-victorops

# MacOC
$ go build

# Ubuntu
$ make build
$ cp $GOPATH/bin/terraform-provider-victorops .
```

Features
------------
Allow Victorops customers to create VictorOps teams, users, escalation policies and routing keys using terraform


Usage
------------
```
provider "victorops" {
  api_id  = "6d700de8"   // An API id tied to an admin user
  api_key = "<REDACTED>" // An API key tied to an admin user
}

// Define the first tf-configured user, 'John Dane'
resource "victorops_user" "jdane_tf" {
  first_name       = "John"
  last_name        = "Dane"
  user_name        = "jdane"
  email            = "jdane51@victorops.com"
  is_admin         = true
}

// Create a new team
resource "victorops_team" "team_vikings" {
  name = "VO-Vikings"
}

// Assigning an existing user to a team
resource "victorops_team_membership" "jdane_membership" {
  team_id          = victorops_team.team_vikings.id
  user_name        = victorops_user.jdane_tf.user_name
}

// Create escalation policies for existing VO rotation (created using portal)
// Note: You need to fetch an existing Rotation Group Slug for using the Escalation Policy
resource "victorops_escalation_policy" "high_severity" {
  name    = "High Severity"
  team_id = victorops_team.team_vikings.id
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

// Create routing keys to push alerts to our escalation policies
resource "victorops_routing_key" "viking_high_severity" {
  name = "viking-high-severity"
  targets = [victorops_escalation_policy.high_severity.id]
}
```

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, build the provider as mentioned above. In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

```sh
$ make testacc
```
