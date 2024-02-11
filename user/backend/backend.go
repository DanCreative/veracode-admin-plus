package backend

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/DanCreative/veracode-admin-plus/user"
	"github.com/DanCreative/veracode-go/veracode"
)

type BackendRepository struct {
	client *veracode.Client
}

func NewBasicBackendRepository(region string) (*BackendRepository, error) {
	key, secret, err := veracode.LoadVeracodeCredentials()
	if err != nil {
		return nil, err
	}

	rateTransport, err := veracode.NewRateTransport(nil, 2/time.Second, 10)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: rateTransport,
		Jar:       jar,
	}

	var regionInt veracode.Region
	switch region {
	case "eu":
		regionInt = veracode.RegionEurope
	case "com":
		regionInt = veracode.RegionCommercial
	case "us":
		regionInt = veracode.RegionUnitedStates
	}

	client, err := veracode.NewClient(regionInt, httpClient, key, secret)
	if err != nil {
		return nil, err
	}

	return &BackendRepository{
		client: client,
	}, nil
}

// SearchAggregatedUsers returns a list of users with each of their roles
func (br *BackendRepository) SearchAggregatedUsers(ctx context.Context, options user.SearchUserOptions) ([]user.User, user.PageMeta, error) {
	summaryUsers, resp, err := br.client.Identity.SearchUsers(ctx, veracode.SearchUserOptions{
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
	})

	if err != nil {
		return nil, user.PageMeta{}, err
	}

	userOrder := make(map[string]int, len(summaryUsers))
	aggregatedUsers := make([]user.User, len(summaryUsers))

	for k, v := range summaryUsers {
		userOrder[v.UserId] = k
	}

	userMap := make(map[string]*user.User)

	var wg sync.WaitGroup

	for _, user := range summaryUsers {
		wg.Add(1)
		go func(userId string) {
			defer wg.Done()
			user, _, err := br.client.Identity.GetUser(ctx, userId, true)
			if err != nil {
				return
			}
			userMap[userId] = veracodeToUser(user)
		}(user.UserId)
	}

	wg.Wait()

	for _, user := range userMap {
		aggregatedUsers[userOrder[user.UserId]] = *user
	}

	return aggregatedUsers, user.PageMeta{
		PageNumber:    resp.Page.Number,
		Size:          resp.Page.Size,
		TotalElements: resp.Page.TotalElements,
		TotalPages:    resp.Page.TotalPages,
		First:         resp.Links.First.HrefURL,
		Last:          resp.Links.Last.HrefURL,
		Next:          resp.Links.Next.HrefURL,
		Prev:          resp.Links.Prev.HrefURL,
		Self:          resp.Links.Self.HrefURL,
	}, nil
}

// BulkUpdateUsers updates multiple users async
func (br *BackendRepository) BulkUpdateUsers(ctx context.Context, users map[string]user.User) []error {
	chError := make(chan error, len(users))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for k, v := range users {
		wg.Add(1)
		go func(userId string, user user.User, ch chan error) {
			vUser := userToVeracode(user, userId)
			p := true
			_, _, err := br.client.Identity.UpdateUser(ctx, vUser, veracode.UpdateOptions{Partial: &p})

			ch <- err
			wg.Done()

		}(k, v, chError)
	}
	go func() {
		wg.Wait()
		close(chError)
	}()

	var errors []error

	for err := range chError {
		if err != nil {
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}
	}
	return errors
}

// GetAllRoles takes a Context and returns a list of user.Role. Currently max page size is set to 100.
func (br *BackendRepository) GetAllRoles(ctx context.Context) ([]user.Role, error) {
	vroles, _, err := br.client.Identity.ListRoles(ctx, veracode.PageOptions{Size: 100})
	if err != nil {
		return nil, err
	}

	droles := make([]user.Role, len(vroles))
	for k, role := range vroles {
		droles[k].IsApi = role.IsApi
		droles[k].IsScanType = role.IsScanType
		droles[k].RoleDescription = role.RoleDescription
		droles[k].RoleId = role.RoleId
		droles[k].RoleName = role.RoleName
	}
	return droles, nil
}

// GetAllTeams takes a Context and returns a list of user.Team. Currently max page size is set to 100.
func (br *BackendRepository) GetAllTeams(ctx context.Context) ([]user.Team, error) {
	vteams, _, err := br.client.Identity.ListTeams(ctx, veracode.ListTeamOptions{Size: 100})
	if err != nil {
		return nil, err
	}

	dteams := make([]user.Team, len(vteams))
	for k, team := range vteams {
		dteams[k].TeamId = team.TeamId
		dteams[k].TeamLegacyId = team.TeamLegacyId
		dteams[k].TeamName = team.TeamName
		dteams[k].Relationship = team.Relationship.Name
	}

	return dteams, nil
}

func userToVeracode(user user.User, userId string) *veracode.User {
	vRoles := make([]veracode.Role, len(user.Roles))
	vTeams := make([]veracode.Team, len(user.Teams))

	for k, role := range user.Roles {
		vRoles[k].RoleId = role.RoleId
	}

	for k, team := range user.Teams {
		vTeams[k].TeamId = team.TeamId
		vTeams[k].Relationship.Name = team.Relationship
	}

	return &veracode.User{
		UserId: userId,
		Teams:  &vTeams,
		Roles:  &vRoles,
	}
}

func veracodeToUser(vUser *veracode.User) *user.User {
	uRoles := make([]user.Role, len(*vUser.Roles))
	uTeams := make([]user.Team, len(*vUser.Teams))

	for k, role := range *vUser.Roles {
		uRoles[k].RoleId = role.RoleId
	}

	for k, team := range *vUser.Teams {
		uTeams[k].TeamId = team.TeamId
		uTeams[k].Relationship = team.Relationship.Name
	}

	return &user.User{
		UserId:       vUser.UserId,
		EmailAddress: vUser.EmailAddress,
		AccountType:  vUser.AccountType,
		Teams:        uTeams,
		Roles:        uRoles,
	}
}