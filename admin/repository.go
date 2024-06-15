package admin

import "context"

type IdentityRepository interface {
	SearchAggregatedUsers(ctx context.Context, options SearchUserOptions) ([]User, PageMeta, error)
	BulkUpdateUsers(ctx context.Context, users map[string]User) []error
	UpdateUser(ctx context.Context, userId string, user User) (User, error)
}

// UserEntityRepository will for now be used to get the different child entities of the user model from the backend.
// It should store entities locally to reduce load times.
type UserEntityRepository[t any] interface {
	// List should get entities from the backend or from the locally stored list.
	// shouldRefresh bool should determine whether the data should be refreshed.
	List(ctx context.Context, options PageOptions, shouldRefresh bool) ([]t, error)
}

// RoleRepository should store roles locally to reduce load times.
type RoleRepository interface {
	// GetRoles should get roles from the backend or from the locally stored roles.
	// shouldRefresh bool should determine whether the data should be refreshed.
	GetRoles(ctx context.Context, options PageOptions, shouldRefresh bool) ([]Role, error)
}
