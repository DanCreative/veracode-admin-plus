package models

import "encoding/json"

type ContextKey string

type Result struct {
	Message   string
	IsSuccess bool
}

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
	AccountType         string `json:"account_type,omitempty"`
	EmailAddress        string `json:"email_address,omitempty"`
	Teams               []Team `json:"teams"`
	CountScanTypeAdders int    `json:"-"`
	Altered             bool   `json:"-"`
	//UserName            string `json:"user_name,omitempty"`
}

type Team struct {
	TeamId       string           `json:"team_id,omitempty"`
	TeamLegacyId int              `json:"team_legacy_id,omitempty"`
	TeamName     string           `json:"team_name,omitempty"`
	Relationship TeamRelationship `json:"relationship,omitempty"`
}

type TeamRelationship struct {
	Name string `json:"name,omitempty"`
}

func (t TeamRelationship) MarshalJSON() ([]byte, error) {
	// jsonValue, err := json.Marshal(map[string]interface{}{
	// 	"relationship": t.Name,
	// })
	jsonValue, err := json.Marshal(t.Name)
	return jsonValue, err
}

type PageMeta struct {
	Number        int `json:"number,omitempty"`
	Size          int `json:"size,omitempty"`
	TotalElements int `json:"total_elements,omitempty"`
	TotalPages    int `json:"total_pages,omitempty"`
	FirstElement  int
	LastElement   int
}

type Filter struct {
	FriendlyLabel string
	Label         string
	FriendlyValue string
	Value         string
	CanUpdate     bool   // Can value be changed
	CanDelete     bool   // Parameter can be removed
	DefaultValue  string // Default value set at startup
}
