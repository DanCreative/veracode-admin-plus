package admin

import (
	"context"
	"errors"
	"fmt"
)

var (
	errNoRoles            = errors.New("user requires at least one role")
	errNoTeams            = errors.New("user has a role that requires membership to a team")
	errRolesRequired      = errors.New("user has a role that requires one of the following roles: Administrator, Security Insigths, Security Lead, Executive, Creator, Submitter or Reviewer")
	errScanTypeRequired   = errors.New("user has a role that requires a scan type role")
	errAdminTeamsRequired = errors.New("user has the team admin role that requires admin team relationships")
)

var _ UserService = &service{}

type UserService interface {
	GetAggregatedUsers(ctx context.Context, options SearchUserOptions) ([]User, PageMeta, error)
	GetUser(ctx context.Context, userId string) (User, error)
	GetRoles(ctx context.Context, shouldRefresh bool) ([]Role, error)
	GetTeams(ctx context.Context, shouldRefresh bool) ([]Team, error)
	UpdateUser(ctx context.Context, userId string, roles map[string]Role, teams map[string]Team) error                                      // Part 1 of user edit: Roles and Teams are updated on the page.
	SubmitUser(ctx context.Context, userId string, accountType string, scanTypes map[string]Role, adminTeams map[string]Team) (User, error) // Part 2 of user edit: Scan Types and Team Admin teams will be set on the submission form if information is required.
}

type service struct {
	userRepo  IdentityRepository
	roleRepo  UserEntityRepository[Role]
	teamRepo  UserEntityRepository[Team]
	userCache userStorage
	userCart  userStorage
}

func NewUserService(identityRepo IdentityRepository, roleRepo UserEntityRepository[Role], teamRepo UserEntityRepository[Team]) *service {
	return &service{
		userRepo:  identityRepo,
		roleRepo:  roleRepo,
		teamRepo:  teamRepo,
		userCache: newUserStorage(),
		userCart:  newUserStorage(),
	}
}

// GetAggregatedUsers gets all of the users and their roles and teams
// If cart value is yes, it gets the data from the cart instead of the veracode backend
//
// Updates the cache and splits the user roles and scan type roles.
func (s *service) GetAggregatedUsers(ctx context.Context, options SearchUserOptions) ([]User, PageMeta, error) {
	var users []User
	var pageMeta PageMeta
	var err error

	// If cart was selected, get all cart users, otherwise get users from the backend
	if options.Cart == "Yes" {
		//users, pageMeta, err = s.userLocalRepo.GetCartUsers(ctx, options)
		fmt.Println("Cart functionality not working yet")
	} else {
		users, pageMeta, err = s.userRepo.SearchAggregatedUsers(ctx, options)
	}
	if err != nil {
		return nil, PageMeta{}, err
	}

	// Remove the scan type roles from the user's Roles map and add it to the user's scan types map.
	for _, user := range users {
		s.splitScanRoles(ctx, user)
	}

	// Add users to cache
	s.userCache.ReplaceUsers(users)

	// Get user from cart and update list of users
	for k, backendUser := range users {
		if cartUser, isFound := s.userCart.GetUserWithID(backendUser.UserId); isFound {
			users[k] = cartUser
		}
	}

	return users, pageMeta, nil
}

// GetRoles gets all of the roles from the RoleRepository.
// Currently max page size is set to 100.
func (s *service) GetRoles(ctx context.Context, shouldRefresh bool) ([]Role, error) {
	return s.roleRepo.List(ctx, PageOptions{Size: 100}, shouldRefresh)
}

// GetTeams gets all of the teams from the TeamRepository.
// Currently max page size is set to 100.
func (s *service) GetTeams(ctx context.Context, shouldRefresh bool) ([]Team, error) {
	return s.teamRepo.List(ctx, PageOptions{Size: 100}, shouldRefresh)
}

// GetUser returns a User from the backend or the local storage.
// Currently only gets user from the local storage.
func (s *service) GetUser(ctx context.Context, userId string) (User, error) {
	user, isFound := s.userCart.GetUserWithID(userId)
	if isFound {
		return user, nil
	}

	user, isFound = s.userCache.GetUserWithID(userId)
	if isFound {
		return user, nil
	}

	return User{}, fmt.Errorf("could not find user with id %s", userId)
}

// UpdateUser updates the locally cached user, but does not commit the changes to the backend.
//
// Updates the cache.
func (s *service) UpdateUser(ctx context.Context, userId string, roles map[string]Role, teams map[string]Team) error {
	cachedUser, isFound := s.userCache.GetUserWithID(userId)
	if !isFound {
		// This should NEVER run
		return fmt.Errorf("user with id: %s was not found in the local cache", userId)
	}

	if roles == nil || len(roles) < 1 {
		// User requires at least one role
		return errNoRoles
	}

	if requiresOtherRoles(roles) {
		// Certain roles requires that specific other roles are also included.
		return errRolesRequired
	}

	if !requiresScanTypeRoles(roles) {
		// If the user does not have a role that requires scan type roles, clear the scan types.
		// this is required because the Veracode backend will throw an error if unnecessary scan type roles are included.
		clear(cachedUser.ScanTypeRoles)
	}

	if (teams == nil || len(teams) < 1) && requiresTeamMembership(roles) {
		// If the end user removes all teams from a user, this checks whether the user has any roles
		// that requires team membership.
		return errNoTeams
	}

	// Lastly, if the user had the TeamAdmin role, preserve the existing admin relationships (if those teams are in the incoming update).
	if _, ok := cachedUser.Roles["TeamAdmin"]; ok {
		for teamId, team := range cachedUser.Teams {
			if team.Relationship == "ADMIN" {
				if _, ok := teams[teamId]; ok {
					teams[teamId] = team
				}
			}
		}
	}

	cachedUser.Roles = roles
	cachedUser.Teams = teams

	s.userCache.AddUser(userId, cachedUser)
	return nil
}

// SubmitUser takes a map of the teams that the user is admin for and a map of the user's scan types, and does the following:
//   - Validates that scan type roles were set if the user requires them.
//   - Validates that admin teams were provided if the user has the teamAdmin role.
//   - Adds the user admin membership to all of their teams.
//   - Updates the cache and splits the user roles and scan type roles.
func (s *service) SubmitUser(ctx context.Context, userId string, accountType string, scanTypes map[string]Role, adminTeams map[string]Team) (User, error) {
	cachedUser, isFound := s.userCache.GetUserWithID(userId)
	if !isFound {
		// This should NEVER run
		return User{}, fmt.Errorf("user with id: %s was not found in the local cache", userId)
	}

	if (scanTypes == nil || len(scanTypes) < 1) && requiresScanTypeRoles(cachedUser.Roles) {
		// if the user has a role that requires a scan type role and the provided scanTypes is empty.
		// Frontend should prevent this block from running.
		return User{}, errScanTypeRequired
	}

	if _, ok := cachedUser.Roles["teamAdmin"]; (adminTeams == nil || len(adminTeams) < 1) && ok {
		// If the user has the teamAdmin roles they should also have teams that they govern.
		// Frontend should prevent this block from running.
		return User{}, errAdminTeamsRequired
	}

	for k := range adminTeams {
		if userTeam, ok := cachedUser.Teams[k]; ok {
			userTeam.Relationship = "ADMIN"
			cachedUser.Teams[k] = userTeam
		}
	}

	cachedUser.ScanTypeRoles = scanTypes

	updatedUser, err := s.userRepo.UpdateUser(ctx, userId, cachedUser)
	if err != nil {
		return User{}, err
	}

	// Account type field is not on the returned user after updating user.
	updatedUser.AccountType = accountType

	// Remove the scan type roles from the user's Roles map and add it to the user's scan types map.
	s.splitScanRoles(ctx, updatedUser)

	// Add the updated user back to the cache
	s.userCache.AddUser(userId, updatedUser)
	return updatedUser, nil
}

// splitScanRoles is a helper method that removes scan type roles from the user's roles field and adds them to the user's scan roles field.
// splitScanRoles calls the service's GetRoles() method, but does not force a refresh if it is not neccessary.
func (s *service) splitScanRoles(ctx context.Context, user User) {
	roles, _ := s.GetRoles(ctx, false)
	for _, role := range roles {
		if role.IsScanType {
			if _, ok := user.Roles[role.RoleName]; ok {
				user.ScanTypeRoles[role.RoleName] = role
				delete(user.Roles, role.RoleName)
			}
		}
	}
}
