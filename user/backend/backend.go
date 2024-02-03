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

	rateTransport, err := veracode.NewRateTransport(nil, time.Second, 10)
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

// TODO: Implement body
func (br *BackendRepository) SearchAggregatedUsers(ctx context.Context, options user.SearchUserOptions) ([]user.User, user.PageMeta, error) {
	return nil, user.PageMeta{}, nil
}

// BulkUpdateUsers updates multiple users async
func (br *BackendRepository) BulkUpdateUsers(ctx context.Context, users map[string]user.User) []error {
	// logrus.WithFields(logrus.Fields{"Function": "BulkPutPartialUsers"}).Trace("Start")
	chError := make(chan error, len(users))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for k, v := range users {
		wg.Add(1)
		go func(userId string, user user.User, ch chan error) {
			vUser := UserToVeracode(user)
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
			// logrus.Errorf("Error picked up during bulk put: %s", err)
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

func UserToVeracode(user user.User) *veracode.User {
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
		UserId: user.UserId,
		Teams:  &vTeams,
		Roles:  &vRoles,
	}
}
