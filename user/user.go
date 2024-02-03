package user

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
	First         string
	Last          string
	Next          string
	Prev          string
	Self          string
}

type SearchUserOptions struct {
	Detailed     string // Passing detailed will return additional hidden fields. Value should be one of: Yes or No
	Page         int    // Page through the list.
	Size         int    // Increase the page size.
	SearchTerm   string // You can search for partial strings of the username, first name, last name, or email address.
	RoleId       string // Filter users by their role. Value should be a valid Role Id.
	UserType     string // Filter by user type. Value should be one of: user or api
	LoginEnabled string // Filter by whether the login is enabled. Value should be one of: Yes or No
	LoginStatus  string // Filter by the login status. Value should be one of: Active, Locked or Never
	SamlUser     string // Filter by whether the user is a SAML user or not. Value should be one of: Yes or No
	TeamId       string // Filter users by team membership. Value should be a valid Team Id.
	ApiId        string // Filter user by their API Id.
}

type UpdateOptions struct {
	Incremental *bool // incremental=true indicates that you are adding items to a list for an object property, such as adding users to a team.
	Partial     *bool // partial=true indicates that you are updating only a subset of properties for an object.
}
