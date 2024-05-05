package admin

import (
	"net/url"
)

type Role struct {
	RoleId          string
	RoleName        string
	RoleDescription string
	IsApi           bool
	IsScanType      bool

	// Not on Veracode API Model
	IsChecked      bool
	IsDisabled     bool
	IsAddScanTypes bool
}

type User struct {
	Roles        []Role
	UserId       string
	AccountType  string
	EmailAddress string
	Teams        []Team

	// Not on Veracode API Model
	CountScanTypeAdders int
	Altered             bool
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
	FirstParams   string // URL parameters for the first page
	LastParams    string // URL parameters for the last page
	NextParams    string // URL parameters for the next page
	PrevParams    string // URL parameters for the prev page
	SelfParams    string // URL parameters for the current page
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

func NewPageMeta(number, size, totalElements, totalPages int, firstUrlStr, lastUrlStr, nextUrlStr, prevUrlStr, selfUrlStr string) PageMeta {
	firstUrl, _ := url.Parse(firstUrlStr)
	lastUrl, _ := url.Parse(lastUrlStr)
	nextUrl, _ := url.Parse(nextUrlStr)
	prevUrl, _ := url.Parse(prevUrlStr)
	selfUrl, _ := url.Parse(selfUrlStr)

	return PageMeta{
		PageNumber:    number,
		Size:          size,
		TotalElements: totalElements,
		TotalPages:    totalPages,
		FirstParams:   firstUrl.RawQuery,
		LastParams:    lastUrl.RawQuery,
		NextParams:    nextUrl.RawQuery,
		PrevParams:    prevUrl.RawQuery,
		SelfParams:    selfUrl.RawQuery,
	}
}

// HasRole is a helper method to check whether a user has a specific role based on name.
func (u *User) HasRole(roleName string) bool {
	for _, userRole := range u.Roles {
		if roleName == userRole.RoleName {
			return true
		}
	}
	return false
}

// HasTeam is a helper method to check whether a user is part of a specific team based on id.
func (u *User) HasTeam(teamId string) bool {
	for _, userTeam := range u.Teams {
		if teamId == userTeam.TeamId {
			return true
		}
	}
	return false
}
