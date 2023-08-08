package models

type Role struct {
	RoleId          string `json:"role_id,omitempty"`
	RoleName        string `json:"role_name,omitempty"`
	RoleDescription string `json:"role_description,omitempty"`
	IsApi           bool   `json:"is_api,omitempty"`
	IsScanType      bool   `json:"is_scan_type,omitempty"`
	IsChecked       bool   `json:"-"`
	IsDisabled      bool   `json:"-"`
	IsAddScanTypes  bool   `json:"-"`
}

type User struct {
	Roles               []Role `json:"roles,omitempty"`
	UserId              string `json:"user_id,omitempty"`
	UserName            string `json:"user_name,omitempty"`
	AccountType         string `json:"account_type,omitempty"`
	EmailAddress        string `json:"email_address,omitempty"`
	Teams               []Team `json:"teams"`
	CountScanTypeAdders int    `json:"-"`
}

type Team struct {
	TeamId       string `json:"team_id,omitempty"`
	TeamLegacyId int    `json:"team_legacy_id,omitempty"`
	TeamName     string `json:"team_name,omitempty"`
}

type PageMeta struct {
	Number        int `json:"number,omitempty"`
	Size          int `json:"size,omitempty"`
	TotalElements int `json:"total_elements,omitempty"`
	TotalPages    int `json:"total_pages,omitempty"`
	FirstElement  int
	LastElement   int
}
