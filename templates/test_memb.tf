// Define the first tf-configured user, 'John Dane'
resource "victorops_user" "test_user" {
  first_name       = "{{.User.FirstName}}"
  last_name        = "{{.User.LastName}}"
  user_name        = "{{.User.Username}}"
  email            = "{{.User.Email}}"
  is_admin         = true
  replacement_user = "{{.Replacement}}"
}

// Create a new team
resource "victorops_team" "test_team" {
  name = "{{.TeamName}}"
}

// Assigning an existing user to a team
resource "victorops_team_membership" "test_membership" {
  team_id          = victorops_team.test_team.id
  user_name        = victorops_user.test_user.user_name
}