package demo

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/DanCreative/veracode-admin-plus/admin"
	"github.com/google/go-querystring/query"
)

type searchUserOptions struct {
	Detailed     string `url:"detailed,omitempty"`      // Passing detailed will return additional hidden fields. Value should be one of: Yes or No
	Page         int    `url:"page"`                    // Page through the list.
	Size         int    `url:"size"`                    // Increase the page size.
	SearchTerm   string `url:"search_term,omitempty"`   // You can search for partial strings of the username, first name, last name, or email address.
	RoleId       string `url:"role_id,omitempty"`       // Filter users by their role. Value should be a valid Role Id.
	UserType     string `url:"user_type,omitempty"`     // Filter by user type. Value should be one of: user or api
	LoginEnabled string `url:"login_enabled,omitempty"` // Filter by whether the login is enabled. Value should be one of: Yes or No
	LoginStatus  string `url:"login_status,omitempty"`  // Filter by the login status. Value should be one of: Active, Locked or Never
	SamlUser     string `url:"saml_user,omitempty"`     // Filter by whether the user is a SAML user or not. Value should be one of: Yes or No
	TeamId       string `url:"team_id,omitempty"`       // Filter users by team membership. Value should be a valid Team Id.
	ApiId        string `url:"api_id,omitempty"`        // Filter user by their API Id.
}

type UserRole struct {
	RoleId          string `json:"role_id"`
	RoleName        string `json:"role_name"`
	RoleDescription string `json:"role_description"`
}

type role struct {
	IsApi           bool   `json:"is_api"`
	IsScanType      bool   `json:"is_scan_type"`
	RoleDescription string `json:"role_description"`
	RoleId          string `json:"role_id"`
	RoleName        string `json:"role_name"`
}

type User struct {
	UserId       string     `json:"user_id"`
	AccountType  string     `json:"account_type"`
	EmailAddress string     `json:"email_address"`
	Roles        []UserRole `json:"roles"`
	Teams        []Team     `json:"teams"`
}

type Team struct {
	TeamId       string `json:"team_id"`
	TeamLegacyId int    `json:"team_legacy_id"`
	TeamName     string `json:"team_name"`
	Relationship string `json:"relationship"`
}

var _ admin.IdentityRepository = &DemoUserRepository{}
var _ admin.UserEntityRepository[admin.Team] = &DemoTeamRepository{}
var _ admin.UserEntityRepository[admin.Role] = &DemoRoleRepository{}

type DemoUserRepository struct {
	demoDataFolder string
}

type DemoTeamRepository struct {
	demoDataFolder string
}

type DemoRoleRepository struct {
	demoDataFolder string
}

func NewDemoUserRepository(demoDataFolder string) *DemoUserRepository {
	return &DemoUserRepository{
		demoDataFolder: demoDataFolder,
	}
}

func NewDemoTeamRepository(demoDataFolder string) *DemoTeamRepository {
	return &DemoTeamRepository{
		demoDataFolder: demoDataFolder,
	}
}

func NewDemoRoleRepository(demoDataFolder string) *DemoRoleRepository {
	return &DemoRoleRepository{
		demoDataFolder: demoDataFolder,
	}
}

func (d *DemoUserRepository) SearchAggregatedUsers(ctx context.Context, options admin.SearchUserOptions) ([]admin.User, admin.PageMeta, error) {
	time.Sleep(200 * time.Millisecond)
	searchOptions := searchUserOptions{
		Detailed:     options.Detailed,
		Page:         options.Page,
		Size:         options.Size,
		SearchTerm:   options.SearchTerm,
		RoleId:       options.RoleId,
		UserType:     options.UserType,
		LoginEnabled: options.LoginEnabled,
		LoginStatus:  options.LoginStatus,
		SamlUser:     options.SamlUser,
		TeamId:       options.TeamId,
		ApiId:        options.ApiId,
	}

	f, err := os.OpenFile(path.Join(d.demoDataFolder, "users.json"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, admin.PageMeta{}, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	domainUsers := make([]admin.User, 0)

	// Read open bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, admin.PageMeta{}, err
	}

	for decoder.More() {
		var user User

		err := decoder.Decode(&user)
		if err != nil {
			return nil, admin.PageMeta{}, err
		}

		if domainUser := userToDomain(user); isUserMatched(domainUser, searchOptions) {
			domainUsers = append(domainUsers, domainUser)
		}
	}

	// Read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, admin.PageMeta{}, err
	}

	// Pagination
	totalElements := len(domainUsers)
	totalPages := int(math.Ceil(float64(totalElements) / float64(searchOptions.Size)))
	paginatedUser := make([]admin.User, 0)

	for i := (searchOptions.Page * searchOptions.Size); i < int(math.Min(float64(totalElements), float64(searchOptions.Size*(searchOptions.Page+1)))); i++ {
		paginatedUser = append(paginatedUser, domainUsers[i])
	}

	// Navigation
	var firstUrlStr, lastUrlStr, nextUrlStr, prevUrlStr, selfUrlStr string

	selfUrl, _ := query.Values(searchOptions)
	selfUrlStr = "/users?" + selfUrl.Encode()

	searchOptions.Page = 0
	firstUrl, _ := query.Values(searchOptions)
	firstUrlStr = "/users?" + firstUrl.Encode()

	searchOptions.Page = totalPages - 1
	lastUrl, _ := query.Values(searchOptions)
	lastUrlStr = "/users?" + lastUrl.Encode()

	if options.Page < totalPages-1 {
		searchOptions.Page = options.Page + 1
		nextUrl, _ := query.Values(searchOptions)
		nextUrlStr = "/users?" + nextUrl.Encode()
	}

	if options.Page > 0 {
		searchOptions.Page = options.Page - 1
		prevUrl, _ := query.Values(searchOptions)
		prevUrlStr = "/users?" + prevUrl.Encode()
	}

	return paginatedUser, admin.NewPageMeta(
		options.Page,
		searchOptions.Size,
		totalElements,
		totalPages,
		firstUrlStr,
		lastUrlStr,
		nextUrlStr,
		prevUrlStr,
		selfUrlStr,
	), nil
}

func (d *DemoUserRepository) UpdateUser(ctx context.Context, userId string, user admin.User) (admin.User, error) {
	time.Sleep(200 * time.Millisecond)
	ruser := admin.User{
		UserId:        user.UserId,
		EmailAddress:  user.EmailAddress,
		Teams:         user.Teams,
		Roles:         make(map[string]admin.Role, len(user.Roles)+len(user.ScanTypeRoles)),
		ScanTypeRoles: make(map[string]admin.Role),
	}

	for roleName, role := range user.Roles {
		ruser.Roles[roleName] = role
	}
	for roleName, role := range user.ScanTypeRoles {
		ruser.Roles[roleName] = role
	}

	return ruser, nil
}

func (d *DemoRoleRepository) List(ctx context.Context, options admin.PageOptions, shouldRefresh bool) ([]admin.Role, error) {
	f, err := os.OpenFile(path.Join(d.demoDataFolder, "roles.json"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	roles := make([]admin.Role, 0)

	// Read open bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}

	for decoder.More() {
		var role role

		err := decoder.Decode(&role)
		if err != nil {
			return nil, err
		}

		roles = append(roles, admin.Role{
			RoleId:          role.RoleId,
			RoleName:        role.RoleName,
			RoleDescription: role.RoleDescription,
			IsApi:           role.IsApi,
			IsScanType:      role.IsScanType,
		})
	}

	// Read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (d *DemoTeamRepository) List(ctx context.Context, options admin.PageOptions, shouldRefresh bool) ([]admin.Team, error) {
	f, err := os.OpenFile(path.Join(d.demoDataFolder, "teams.json"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	teams := make([]admin.Team, 0)

	// Read open bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}

	for decoder.More() {
		var team Team

		err := decoder.Decode(&team)
		if err != nil {
			return nil, err
		}

		teams = append(teams, admin.Team{
			TeamId:   team.TeamId,
			TeamName: team.TeamName,
		})
	}

	// Read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func userToDomain(demoUser User) admin.User {
	uRoles := make(map[string]admin.Role)
	uScanTypeRoles := make(map[string]admin.Role)
	uTeams := make(map[string]admin.Team)

	for _, role := range demoUser.Roles {
		uRoles[role.RoleName] = admin.Role{
			RoleId:          role.RoleId,
			RoleName:        role.RoleName,
			RoleDescription: role.RoleDescription,
		}
	}

	for _, team := range demoUser.Teams {
		uTeams[team.TeamId] = admin.Team{
			TeamId:       team.TeamId,
			TeamLegacyId: team.TeamLegacyId,
			TeamName:     team.TeamName,
			Relationship: team.Relationship,
		}
	}

	return admin.User{
		UserId:        demoUser.UserId,
		EmailAddress:  demoUser.EmailAddress,
		AccountType:   demoUser.AccountType,
		Teams:         uTeams,
		Roles:         uRoles,
		ScanTypeRoles: uScanTypeRoles,
	}
}

func isUserMatched(domainUser admin.User, options searchUserOptions) bool {
	isMatch := true
	if options.SearchTerm != "" {
		if !strings.Contains(domainUser.EmailAddress, options.SearchTerm) {
			isMatch = false
		}
	}

	if options.TeamId != "" {
		if _, ok := domainUser.Teams[options.TeamId]; !ok {
			isMatch = false
		}
	}

	if options.UserType != "" {
		if strings.ToUpper(options.UserType) != domainUser.AccountType {
			isMatch = false
		}
	}

	return isMatch
}
