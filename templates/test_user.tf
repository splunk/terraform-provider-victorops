resource "victorops_user" "test_user" {
  first_name       = "{{.User.FirstName}}"
  last_name        = "{{.User.LastName}}"
  user_name        = "{{.User.Username}}"
  email            = "{{.User.Email}}"
  is_admin         = true
  replacement_user = "{{.Replacement}}"
}
