package models

type Role struct {
	RoleId          string `json:"role_id"`
	RoleName        string `json:"role_name"`
	RoleDescription string `json:"role_description"`
	IsApi           bool   `json:"is_api"`
	IsScanType      bool   `json:"is_scan_type"`
	IsChecked       bool   `json:"-"`
	IsDisabled      bool   `json:"-"`
	IsAddScanTypes  bool   `json:"-"`
}

type User struct {
	Roles               []Role `json:"roles"`
	UserId              string `json:"user_id"`
	UserName            string `json:"user_name"`
	AccountType         string `json:"account_type"`
	EmailAddress        string `json:"email_address"`
	Teams               []Team `json:"teams"`
	CountScanTypeAdders int    `json:"-"`
}

type Team struct {
	TeamId       string `json:"team_id,omitempty"`
	TeamLegacyId int    `json:"team_legacy_id,omitempty"`
	TeamName     string `json:"team_name,omitempty"`
}
