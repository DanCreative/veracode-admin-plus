package mem

import (
	"context"
	"errors"
	"math"
	"net/url"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/user"
	"github.com/gorilla/schema"
)

var (
	_ user.IdentityLocalRepository = &UserLocalMemRepository{}

	ErrUserNotFound = errors.New("user not found in cart")
)

type UserLocalMemRepository struct {
	roleCache []user.Role
	teamCache []user.Team
	userCache map[string]user.User
	userCart  map[string]user.User
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

// GetAllRoles gets all local roles
func (ulr *UserLocalMemRepository) GetAllRoles(ctx context.Context) ([]user.Role, error) {
	return ulr.roleCache, nil
}

// GetAllTeams gets all local teams
func (ulr *UserLocalMemRepository) GetAllTeams(ctx context.Context) ([]user.Team, error) {
	return ulr.teamCache, nil
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
// TODO: Sort cart
func (ulr *UserLocalMemRepository) GetCartUsers(ctx context.Context, options user.SearchUserOptions) ([]user.User, user.PageMeta, error) {
	ulr.muCart.Lock()
	defer ulr.muCart.Unlock()

	totalElements := len(ulr.userCart)
	totalPages := int(math.Ceil(float64(totalElements) / float64(options.Size)))

	encoder := schema.NewEncoder()
	//First page params:
	foptions := options
	foptions.Page = 0
	fparams := getParams(foptions, encoder)

	//Last page params:
	loptions := options
	loptions.Page = totalPages - 1
	lparams := getParams(loptions, encoder)

	//Self page params:
	sparams := getParams(options, encoder)

	//Next page params:
	var nparams string
	if options.Page+1 < totalPages {
		noptions := options
		noptions.Page += 1
		nparams = getParams(noptions, encoder)
	}

	//Prev page params:
	var pparams string
	if options.Page-1 > 0 {
		poptions := options
		poptions.Page -= 1
		pparams = getParams(poptions, encoder)
	}

	totalCartUsers := make([]user.User, 0, totalElements)
	for _, value := range ulr.userCart {
		totalCartUsers = append(totalCartUsers, value)
	}

	pagedUsers := make([]user.User, 0, options.Size)
	for i := options.Page * options.Size; i < int(math.Min(float64((options.Page+1)*options.Size), float64(totalElements))); i++ {

		pagedUsers = append(pagedUsers, totalCartUsers[i])
	}

	return pagedUsers, user.PageMeta{
		Size:          options.Size,
		PageNumber:    options.Page,
		TotalElements: totalElements,
		TotalPages:    totalPages,
		FirstParams:   fparams,
		LastParams:    lparams,
		SelfParams:    sparams,
		NextParams:    nparams,
		PrevParams:    pparams,
	}, nil
}

// getParams is a helper function that converts object to url.Values string
func getParams(options user.SearchUserOptions, encoder *schema.Encoder) string {
	vals := url.Values{}
	encoder.Encode(&options, vals)
	return vals.Encode()
}

// Remove user from the cart
func (ulr *UserLocalMemRepository) RemoveCartUser(ctx context.Context, userId string) error {
	ulr.muCart.Lock()
	defer ulr.muCart.Unlock()

	delete(ulr.userCart, userId)
	return nil
}

// Clear the user cart
func (ulr *UserLocalMemRepository) ClearCart(ctx context.Context) error {
	ulr.muCart.Lock()
	defer ulr.muCart.Unlock()

	clear(ulr.userCart)
	return nil
}
