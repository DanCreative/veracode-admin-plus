package admin

import (
	"net/url"
	"sync"
)

var (
	rolesRequireTeamMembership = [...]string{
		"extmitigationapprover",
		"workSpaceEditor",
		"deletescans",
		"extreviewer",
		"analyticsCreator",
		"securityinsightsonly",
		"workSpaceAdmin",
		"sandboxuser",
		"teamAdmin",
		"extcreator",
		"extsubmitter",
	} // Requires Team Membership
	rolesRequireScanType = [...]string{
		"extcreator",
		"extsubmitter",
		"extseclead",
		"uploadapi",
		"submitterapi",
	} // Requires Scan Type
	rolesRequireOtherRoles = [...]string{
		"collectionManager",
		"collectionReviewer",
		"sandboxadmin",
		"extpolicyadmin",
	} // Requires one of the following roles: Administrator, Security Insigths, Security Lead, Executive, Creator, Submitter or Reviewer
)

type PageOptions struct {
	Size int
	Page int
}

type Role struct {
	RoleId          string
	RoleName        string
	RoleDescription string
	IsApi           bool
	IsScanType      bool
}

type User struct {
	Roles         map[string]Role
	ScanTypeRoles map[string]Role
	UserId        string
	AccountType   string
	EmailAddress  string
	Teams         map[string]Team

	Altered bool // Not on Veracode API Model
}

type Team struct {
	TeamId       string
	TeamLegacyId int
	TeamName     string
	Relationship string
}

type PageMeta struct {
	PageNumber    int
	Size          int
	TotalElements int
	TotalPages    int

	// Below attributes are used for navigation on the frontend.
	FirstParams  string            // URL parameters for the first page
	LastParams   string            // URL parameters for the last page
	NextParams   string            // URL parameters for the next page
	PrevParams   string            // URL parameters for the prev page
	SelfParams   string            // URL parameters for the current page
	FilterParams map[string]string // URL parameters for all of the filters.
}

type SearchUserOptions struct {
	Detailed     string `schema:"detailed,omitempty"`      // Passing detailed will return additional hidden fields. Value should be one of: Yes or No
	Page         int    `schema:"page"`                    // Page through the list.
	Size         int    `schema:"size,omitempty"`          // Increase the page size.
	SearchTerm   string `schema:"search_term,omitempty"`   // You can search for partial strings of the username, first name, last name, or email address.
	RoleId       string `schema:"role_id,omitempty"`       // Filter users by their role. Value should be a valid Role Id.
	UserType     string `schema:"user_type,omitempty"`     // Filter by user type. Value should be one of: user or api
	LoginEnabled string `schema:"login_enabled,omitempty"` // Filter by whether the login is enabled. Value should be one of: Yes or No
	LoginStatus  string `schema:"login_status,omitempty"`  // Filter by the login status. Value should be one of: Active, Locked or Never
	SamlUser     string `schema:"saml_user,omitempty"`     // Filter by whether the user is a SAML user or not. Value should be one of: Yes or No
	TeamId       string `schema:"team_id,omitempty"`       // Filter users by team membership. Value should be a valid Team Id.
	ApiId        string `schema:"api_id,omitempty"`        // Filter user by their API Id.
	Cart         string `schema:"cart,omitempty"`          // Not part of options for the backend. Value should be one of: Yes or No
}

type UpdateOptions struct {
	Incremental *bool // incremental=true indicates that you are adding items to a list for an object property, such as adding users to a team.
	Partial     *bool // partial=true indicates that you are updating only a subset of properties for an object.
}

// message is used on the frontend
type message struct {
	IsSuccess  bool
	ShouldShow bool
	Text       string
}

type userStorage struct {
	users map[string]struct {
		index int
		User
	}
	mu sync.RWMutex
}

func NewPageMeta(number, size, totalElements, totalPages int, firstUrlStr, lastUrlStr, nextUrlStr, prevUrlStr, selfUrlStr string) PageMeta {
	firstUrl, _ := url.Parse(firstUrlStr)
	lastUrl, _ := url.Parse(lastUrlStr)
	nextUrl, _ := url.Parse(nextUrlStr)
	prevUrl, _ := url.Parse(prevUrlStr)
	selfUrl, _ := url.Parse(selfUrlStr)

	selfValues := selfUrl.Query()
	filterParams := map[string]string{}

	for key, value := range selfValues {
		if key == "detailed" || key == "page" || key == "size" {
			continue
		}

		// Make sure that there is atleast one item in list
		if len(value) > 0 {
			// Creates a map of filter parameter sets where each set excludes the current parameter.
			// This is for the delete filter button on the frontend.

			// If the user tries to remove the "User Type" filter, it will instead switch between API or UI.
			if key == "user_type" {
				if value[0] == "user" {
					selfValues[key][0] = "api"
				} else if value[0] == "api" {
					selfValues[key][0] = "user"
				}

			} else {
				delete(selfValues, key)
			}
			filterParams[key] = selfValues.Encode()
			selfValues[key] = value
		}
	}

	return PageMeta{
		PageNumber:    number,
		Size:          size,
		TotalElements: totalElements,
		TotalPages:    totalPages,

		FirstParams:  firstUrl.RawQuery,
		LastParams:   lastUrl.RawQuery,
		NextParams:   nextUrl.RawQuery,
		PrevParams:   prevUrl.RawQuery,
		SelfParams:   selfUrl.RawQuery,
		FilterParams: filterParams,
	}
}

// HasRole is a helper method to check whether a user has a specific role based on name.
func (u *User) HasRole(roleName string) bool {
	_, roleOk := u.Roles[roleName]
	return roleOk
}

// HasScanRole is a helper method to check whether a user has a specific scan type role based on name.
func (u *User) HasScanRole(roleName string) bool {
	_, roleOk := u.ScanTypeRoles[roleName]
	return roleOk
}

// GetAnyScanRole is a helper method to get the API or UI scan type roles.
// If the user does not have the any scan role, it returns an empty string.
func (u *User) GetAnyScanRole() string {
	_, hasAPIAny := u.ScanTypeRoles["apisubmitanyscan"]
	_, hasUIAny := u.ScanTypeRoles["extsubmitanyscan"]

	if hasAPIAny {
		return "apisubmitanyscan"
	} else if hasUIAny {
		return "extsubmitanyscan"
	} else {
		return ""
	}
}

func (u *User) anyScanRoleValue() string {
	if u.AccountType == "USER" {
		return "extsubmitanyscan"
	} else {
		return "apisubmitanyscan"
	}
}

// Certain roles require that the user has scan type roles as well. requiresScanTypeRoles will
// return true if the user does and false if not.
func requiresScanTypeRoles(roles map[string]Role) bool {
	for _, roleName := range rolesRequireScanType {
		if _, ok := roles[roleName]; ok {
			return true
		}

	}

	return false
}

// Certain roles requires that the user is part of a team. requiresTeamMembership returns
// a boolean with value true if the user has one of these roles. If the user does, it also
// returns the first offending role.
//
// The rule does not apply if the user is an admin.
func requiresTeamMembership(roles map[string]Role) bool {
	if _, ok := roles["extadmin"]; ok {
		return false
	}
	for _, roleName := range rolesRequireTeamMembership {
		if _, ok := roles[roleName]; ok {
			return true
		}
	}
	return false
}

// Certain roles require that one of the following roles is selected: Administrator, Security Insigths, Security Lead, Executive, Creator, Submitter or Reviewer.
// requiresOtherRoles will return true if the user has one of these roles and false if not.
func requiresOtherRoles(roles map[string]Role) bool {
	for _, roleName := range rolesRequireOtherRoles {
		if _, ok := roles[roleName]; ok {
			for _, requiredRoleName := range [...]string{"extadmin", "securityinsightsonly", "extseclead", "extexecutive", "extcreator", "extsubmitter", "extreviewer"} {
				if _, ok := roles[requiredRoleName]; ok {
					return false
				}
			}

			return true
		}
	}
	return false
}

// HasTeam is a helper method to check whether a user is part of a specific team based on id.
func (u *User) HasTeam(teamId string) bool {
	_, ok := u.Teams[teamId]
	return ok
}

// HasAdminTeam is a helper method to check whether a user has a team with provided id and whether their relationship to that team is administrative.
func (u *User) HasAdminTeam(teamId string) bool {
	team, ok := u.Teams[teamId]
	if !ok {
		return false
	}
	return team.Relationship == "ADMIN"
}

func newUserStorage() userStorage {
	return userStorage{
		users: make(map[string]struct {
			index int
			User
		}),
	}
}

func (us *userStorage) ReplaceUsers(users []User) {
	us.mu.Lock()
	defer us.mu.Unlock()
	clear(us.users)
	for k, user := range users {
		us.users[user.UserId] = struct {
			index int
			User
		}{
			k,
			user,
		}
	}
}

func (us *userStorage) GetUserWithID(userID string) (User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	iUser, ok := us.users[userID]
	return iUser.User, ok
}

func (us *userStorage) GetUserList() []User {
	us.mu.RLock()
	defer us.mu.RUnlock()

	r := make([]User, len(us.users))

	for _, v := range us.users {
		r[v.index] = v.User
	}

	return r
}

func (us *userStorage) AddUser(userID string, user User) {
	us.mu.Lock()
	defer us.mu.Unlock()

	if iUser, ok := us.users[userID]; ok {
		us.users[userID] = struct {
			index int
			User
		}{iUser.index, user}
	} else {
		us.users[userID] = struct {
			index int
			User
		}{len(us.users), user}
	}
}
