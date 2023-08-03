package main

import (
	veracode "github.com/DanCreative/veracode-admin-plus/Veracode"
	"github.com/DanCreative/veracode-admin-plus/models"
)

// RenderValidation adds roles for rendering purposes.
func RenderValidation(user *models.User) {
	var isAdmin bool
	hasScanTypes := 0

	// Add missing roles
	newRoles := make([]models.Role, len(Roles))
outer:
	for i, systemRole := range Roles {
		for _, userRole := range user.Roles {
			if userRole.RoleId == systemRole.RoleId {
				systemRole.IsChecked = true

				// If admin disable
				if userRole.RoleName == "extadmin" {
					systemRole.IsDisabled = true
					isAdmin = true
				}

				// If Creator, Security Lead and Submitter, add scan types
				if veracode.AddScanTypesRoles[userRole.RoleName] {
					hasScanTypes += 1
				}

				newRoles[i] = systemRole
				continue outer
			}
		}
		// If admin disable
		if systemRole.RoleName == "extadmin" {
			systemRole.IsDisabled = true
		}
		newRoles[i] = systemRole
	}

	if isAdmin || hasScanTypes < 1 {
		// If Admin disable Team Admin
		for i := range newRoles {
			if newRoles[i].RoleName == "teamAdmin" && isAdmin {
				newRoles[i].IsDisabled = true
			}
			if newRoles[i].IsScanType && hasScanTypes < 1 {
				newRoles[i].IsDisabled = true
			}
		}
	}

	user.Roles = newRoles
	user.CountScanTypeAdders = hasScanTypes
}
