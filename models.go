package main

type Role struct {
	RoleId          string `json:"role_id"`
	RoleName        string `json:"role_name"`
	RoleDescription string `json:"role_description"`
	IsApi           bool   `json:"is_api"`
	IsChecked       bool   `json:"-"`
	IsDisabled      bool   `json:"-"`
}

type User struct {
	Roles    []Role `json:"roles"`
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
}
