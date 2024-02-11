package user

import (
	"context"
)

var _ UserService = &service{}

type UserService interface {
	GetAggregatedUsers(ctx context.Context, urlValues string) ([]User, PageMeta, error)
	SubmitUsers(ctx context.Context, users []User) error
	GetRoles(ctx context.Context) ([]Role, error)
	GetTeams(ctx context.Context) ([]Team, error)
}

type service struct {
	userBackendRepo IdentityRepository
	userLocalRepo   IdentityLocalRepository
}

// TODO: Implement body
func (s *service) GetAggregatedUsers(ctx context.Context, urlValues string) ([]User, PageMeta, error) {
	return nil, PageMeta{}, nil
}

// TODO: Implement body
func (s *service) SubmitUsers(ctx context.Context, users []User) error {
	return nil
}

// TODO: Implement body
func (s *service) GetRoles(ctx context.Context) ([]Role, error) {
	return nil, nil
}

// TODO: Implement body
func (s *service) GetTeams(ctx context.Context) ([]Team, error) {
	return nil, nil
}

func NewUserService(identityRepo IdentityRepository, identityLocalRepo IdentityLocalRepository) UserService {
	return &service{
		userBackendRepo: identityRepo,
		userLocalRepo:   identityLocalRepo,
	}
}
