package backend

import (
	"context"
	"errors"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/admin"
	"github.com/DanCreative/veracode-go/veracode"
)

var _ admin.IdentityRepository = &BackendRepository{}

type BackendRepository struct {
	getClient func() (*veracode.Client, error)
}

func NewBasicBackendRepository(getClientFunc func() (*veracode.Client, error)) *BackendRepository {
	return &BackendRepository{
		getClient: getClientFunc,
	}
}

// SearchAggregatedUsers returns a list of users with each of their roles
func (br *BackendRepository) SearchAggregatedUsers(ctx context.Context, options admin.SearchUserOptions) ([]admin.User, admin.PageMeta, error) {
	client, err := br.getClient()
	if err != nil {
		return nil, admin.PageMeta{}, err
	}
	summaryUsers, resp, err := client.Identity.SearchUsers(ctx, veracode.SearchUserOptions{
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
		return nil, admin.PageMeta{}, err
	}

	type userResult struct {
		user  admin.User
		index int
		err   error
	}

	var wg sync.WaitGroup
	resultsChan := make(chan userResult)

	for index, summaryUser := range summaryUsers {
		wg.Add(1)
		go func(summaryUser veracode.User, index int) {
			defer wg.Done()
			fullUser, _, err := client.Identity.GetUser(ctx, summaryUser.UserId, true)
			adminUser := veracodeToUser(fullUser)

			resultsChan <- userResult{
				err:   err,
				user:  adminUser,
				index: index,
			}
		}(summaryUser, index)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	aggregatedUsers := make([]admin.User, len(summaryUsers))
	var finalErr error

	for res := range resultsChan {
		if res.err != nil {
			if finalErr == nil {
				finalErr = res.err
			} else {
				finalErr = errors.Join(finalErr, res.err)
			}
			continue
		}

		aggregatedUsers[res.index] = res.user
	}

	return aggregatedUsers, admin.NewPageMeta(
		resp.Page.Number,
		resp.Page.Size,
		resp.Page.TotalElements,
		resp.Page.TotalPages,
		resp.Links.First.HrefURL,
		resp.Links.Last.HrefURL,
		resp.Links.Next.HrefURL,
		resp.Links.Prev.HrefURL,
		resp.Links.Self.HrefURL), finalErr
}

func (br *BackendRepository) UpdateUser(ctx context.Context, userId string, user admin.User) (admin.User, error) {
	client, err := br.getClient()
	if err != nil {
		return admin.User{}, err
	}

	vUser := userToVeracode(user, userId)
	p := true
	updatedUser, _, err := client.Identity.UpdateUser(ctx, vUser, veracode.UpdateOptions{Partial: &p})
	if err != nil {
		return admin.User{}, err
	}

	return veracodeToUser(updatedUser), nil
}

func userToVeracode(user admin.User, userId string) *veracode.User {
	vRoles := make([]veracode.RoleUser, 0, len(user.Roles)+len(user.ScanTypeRoles))
	vTeams := make([]veracode.Team, 0, len(user.Teams))

	for _, role := range user.Roles {
		vRoles = append(vRoles, veracode.RoleUser{RoleName: role.RoleName})
	}

	for _, role := range user.ScanTypeRoles {
		vRoles = append(vRoles, veracode.RoleUser{RoleName: role.RoleName})
	}

	for _, team := range user.Teams {
		vTeams = append(vTeams, veracode.Team{
			TeamId: team.TeamId,
			Relationship: veracode.TeamRelationship{
				Name: team.Relationship,
			},
		})
	}

	return &veracode.User{
		UserId: userId,
		Teams:  &vTeams,
		Roles:  &vRoles,
	}
}

func veracodeToUser(vUser *veracode.User) admin.User {
	uRoles := make(map[string]admin.Role)
	uScanTypeRoles := make(map[string]admin.Role)
	uTeams := make(map[string]admin.Team)

	for _, role := range *vUser.Roles {
		uRoles[role.RoleName] = admin.Role{
			RoleId:          role.RoleId,
			RoleName:        role.RoleName,
			RoleDescription: role.RoleDescription,
		}
	}

	for _, team := range *vUser.Teams {
		uTeams[team.TeamId] = admin.Team{
			TeamId:       team.TeamId,
			TeamLegacyId: team.TeamLegacyId,
			TeamName:     team.TeamName,
			Relationship: team.Relationship.Name,
		}
	}

	return admin.User{
		UserId:        vUser.UserId,
		EmailAddress:  vUser.EmailAddress,
		AccountType:   vUser.AccountType,
		Teams:         uTeams,
		Roles:         uRoles,
		ScanTypeRoles: uScanTypeRoles,
	}
}
