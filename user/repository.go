package user

import "context"

type IdentityRepository interface {
	SearchAggregatedUsers(ctx context.Context, options SearchUserOptions) ([]User, PageMeta, error)
	BulkUpdateUsers(ctx context.Context, users map[string]User) []error
	GetAllRoles(ctx context.Context) ([]Role, error)
	GetAllTeams(ctx context.Context) ([]Team, error)
}

type UserLocalRepository interface {
	// Add users to local cache
	AddUsers(ctx context.Context, users []User) error

	// Update cached user and move it to the cart
	UpdateUser(ctx context.Context, userId string, user User) error

	// Get all users that have been modified and are in the cart
	GetCartUsers(ctx context.Context) ([]User, error)

	// Remove user from the cart
	RemoveCartUser(ctx context.Context, userId string) error

	// Clear the user cart
	ClearCart(ctx context.Context) error

	// Get user from either cache or cart
	GetUser(ctx context.Context, userId string) (User, error)
}