package user

import (
	"context"
)

type UserService interface {
	GetAggregatedUsers(ctx context.Context, urlValues string) ([]User, PageMeta, error)
	SubmitUsers(ctx context.Context, users []User) error
	GetRoles(ctx context.Context) ([]Role, error)
	GetTeams(ctx context.Context) ([]Team, error)
}

type service struct {
	userBackendRepo IdentityRepository
	userLocalRepo   UserLocalRepository
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

func NewUserService(identityRepo IdentityRepository, localUserRepo UserLocalRepository) UserService {
	return &service{
		userBackendRepo: identityRepo,
		userLocalRepo:   localUserRepo,
	}
}
