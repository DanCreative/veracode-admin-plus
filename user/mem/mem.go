package mem

import (
	"context"
	"errors"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/user"
)

var (
	ErrUserNotFound = errors.New("user not found in cart")
)

type UserLocalMemRepository struct {
	userCart  map[string]user.User
	userCache map[string]user.User
	muCache   sync.Mutex
	muCart    sync.Mutex
}

// Add users to local cache
func (ulr *UserLocalMemRepository) AddUsers(ctx context.Context, users []user.User) error {
	ulr.muCache.Lock()
	defer ulr.muCache.Unlock()

	for _, user := range users {
		ulr.userCache[user.UserId] = user
	}

	return nil
}

// GetUser gets a user with id from the cart
func (ulr *UserLocalMemRepository) GetUser(ctx context.Context, userId string) (user.User, error) {
	ulr.muCart.Lock()
	defer ulr.muCart.Unlock()

	if val, ok := ulr.userCart[userId]; ok {
		return val, nil
	}
	return user.User{}, ErrUserNotFound
}

// Update cached user and move it to the cart
func (ulr *UserLocalMemRepository) UpdateUser(ctx context.Context, userId string, dUser user.User) error {
	cachedUser, err := ulr.GetUser(ctx, userId)
	if err != nil {
		return err
	}
	cachedUser.Teams = dUser.Teams
	cachedUser.Roles = dUser.Roles

	ulr.muCart.Lock()
	defer ulr.muCache.Unlock()

	ulr.userCart[cachedUser.UserId] = cachedUser
	return nil
}

// Get all users that have been modified and are in the cart
// TODO: page meta calculations
func (ulr *UserLocalMemRepository) GetCartUsers(ctx context.Context, options user.SearchUserOptions) ([]user.User, user.PageMeta, error) {
	ulr.muCart.Lock()
	defer ulr.muCart.Unlock()

	rUsers := make([]user.User, len(ulr.userCart))
	for _, value := range ulr.userCart {
		rUsers = append(rUsers, value)
	}

	return rUsers, user.PageMeta{}, nil
}

// Remove user from the cart
func (ulr *UserLocalMemRepository) RemoveCartUser(ctx context.Context, userId string) error {
	return nil
}

// Clear the user cart
func (ulr *UserLocalMemRepository) ClearCart(ctx context.Context) error {
	return nil
}
