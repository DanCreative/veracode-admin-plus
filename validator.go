package main

import "github.com/DanCreative/veracode-admin-plus/models"

// RenderValidation adds roles for rendering purposes.
func RenderValidation(user *models.User) {
	var isAdmin bool
	// Add missing roles
	newRoles := make([]models.Role, len(Roles))
outer:
	for i, systemRole := range Roles {
		for _, userRole := range user.Roles {
			if userRole.RoleId == Roles[i].RoleId {
				userRole.IsChecked = true

				// If admin disable
				if userRole.RoleName == "extadmin" {
					userRole.IsDisabled = true
					isAdmin = true
				}

				// If Creator, Security Lead and Submitter, add scan types
				if userRole.RoleName == "extcreator" || userRole.RoleName == "extseclead" || userRole.RoleName == "extsubmitter" {
					userRole.IsAddScanTypes = true
				}

				newRoles[i] = userRole
				continue outer
			}
		}
		// If admin disable
		if systemRole.RoleName == "extadmin" {
			systemRole.IsDisabled = true
		}
		newRoles[i] = systemRole
	}

	if isAdmin {
		// If Admin disable Team Admin
		for i := range newRoles {
			if newRoles[i].RoleName == "teamAdmin" {
				newRoles[i].IsDisabled = true
			}
		}
	}

	user.Roles = newRoles
}
