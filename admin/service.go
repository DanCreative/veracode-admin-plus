package admin

import (
	"context"
)

var _ UserService = &service{}

type UserService interface {
	GetAggregatedUsers(ctx context.Context, options SearchUserOptions) ([]User, PageMeta, error)
	SubmitUsers(ctx context.Context) []error
	SetRoles(ctx context.Context) error
	GetRoles(ctx context.Context) ([]Role, error)
	GetTeams(ctx context.Context) ([]Team, error)
}

type service struct {
	userBackendRepo IdentityRepository
	userLocalRepo   IdentityLocalRepository
}

// GetAggregatedUsers gets all of the users and their roles and teams
// If cart value is yes, it gets the data from the cart instead of the veracode backend
func (s *service) GetAggregatedUsers(ctx context.Context, options SearchUserOptions) ([]User, PageMeta, error) {
	var users []User
	var pageMeta PageMeta
	var err error

	// If cart was selected, get all cart users, otherwise get users from the backend
	if options.Cart == "Yes" {
		users, pageMeta, err = s.userLocalRepo.GetCartUsers(ctx, options)
	} else {
		users, pageMeta, err = s.userBackendRepo.SearchAggregatedUsers(ctx, options)
	}
	if err != nil {
		return nil, PageMeta{}, err
	}

	// Add users to cache
	err = s.userLocalRepo.AddUsers(ctx, users)
	if err != nil {
		return nil, PageMeta{}, err
	}

	// Get user from cart and update list of users
	for k, backendUser := range users {
		if cartUser, isFound := s.userLocalRepo.GetUser(ctx, backendUser.UserId); isFound {
			users[k] = cartUser
			users[k].Altered = true
		}
	}

	return users, pageMeta, nil
}

// SubmitUsers bulk updates all of the users in the cart.
func (s *service) SubmitUsers(ctx context.Context) []error {
	users, _, err := s.userLocalRepo.GetCartUsers(ctx, SearchUserOptions{Cart: "Yes"})
	if err != nil {
		return []error{err}
	}

	updateUsers := make(map[string]User, len(users))
	for _, user := range users {
		updateUsers[user.UserId] = user
	}

	errs := s.userBackendRepo.BulkUpdateUsers(ctx, updateUsers)
	if errs != nil {
		return errs
	}

	return nil
}

// GetRoles gets all of the roles from the local repo.
// Roles should be stored in the local repo upon init.
func (s *service) GetRoles(ctx context.Context) ([]Role, error) {
	return s.userLocalRepo.GetAllRoles(ctx)
}

// SetRoles gets all roles from the backend and caches them
// in the local repository.
func (s *service) SetRoles(ctx context.Context) error {
	roles, err := s.userBackendRepo.GetAllRoles(ctx)
	if err != nil {
		return err
	}

	return s.userLocalRepo.SetRoles(ctx, roles)
}

// GetTeams will always get the teams from the Veracode backend.
func (s *service) GetTeams(ctx context.Context) ([]Team, error) {
	return s.userBackendRepo.GetAllTeams(ctx)
}

func NewUserService(identityRepo IdentityRepository, identityLocalRepo IdentityLocalRepository) *service {
	return &service{
		userBackendRepo: identityRepo,
		userLocalRepo:   identityLocalRepo,
	}
}
