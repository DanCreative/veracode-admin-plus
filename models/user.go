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
	Roles        []Role `json:"roles"`
	UserId       string `json:"user_id"`
	UserName     string `json:"user_name"`
	EmailAddress string `json:"email_address"`
}
