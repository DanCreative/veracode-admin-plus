package main

// AddMissingRoles adds roles for rendering purposes.
func RenderValidation(user *User) {
	// Add missing roles
	newRoles := make([]Role, len(Roles))
outer:
	for i, systemRole := range Roles {
		for _, userRole := range user.Roles {
			if userRole.RoleId == Roles[i].RoleId {
				userRole.IsChecked = true
				newRoles[i] = userRole
				continue outer
			}
		}
		newRoles[i] = systemRole
	}

	user.Roles = newRoles
	// If Admin disable Team Admin

	// If Creator, Security Lead and Submitter, add scan types

	// Sort roles
}
