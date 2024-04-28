package mem

import (
	"context"
	"errors"
	"math"
	"net/url"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/admin"
	"github.com/gorilla/schema"
)

var (
	_ admin.IdentityLocalRepository = &UserLocalMemRepository{}

	ErrUserNotFound = errors.New("user not found in cart")
)

type UserLocalMemRepository struct {
	roleCache []admin.Role
	userCache map[string]admin.User
	userCart  map[string]admin.User
	muCache   sync.Mutex
	muCart    sync.Mutex
}

func NewUserLocalMemRepository() *UserLocalMemRepository {
	return &UserLocalMemRepository{
		roleCache: make([]admin.Role, 0),
		userCache: make(map[string]admin.User),
		userCart:  make(map[string]admin.User),
	}
}

// Add users to local cache
func (ulr *UserLocalMemRepository) AddUsers(ctx context.Context, users []admin.User) error {
	ulr.muCache.Lock()
	defer ulr.muCache.Unlock()

	for _, user := range users {
		ulr.userCache[user.UserId] = user
	}

	return nil
}

// GetUser gets a user with id from the cart.
// GetUser will return 2 values. The first is the object and the second is a bool indicating whether the user was found.
func (ulr *UserLocalMemRepository) GetUser(ctx context.Context, userId string) (admin.User, bool) {
	ulr.muCart.Lock()
	defer ulr.muCart.Unlock()

	if val, ok := ulr.userCart[userId]; ok {
		return val, true
	}
	return admin.User{}, false
}

// GetAllRoles gets all local roles
func (ulr *UserLocalMemRepository) GetAllRoles(ctx context.Context) ([]admin.Role, error) {
	return ulr.roleCache, nil
}

// Update cached user and move it to the cart
func (ulr *UserLocalMemRepository) UpdateUser(ctx context.Context, userId string, dUser admin.User) error {
	cachedUser, isFound := ulr.GetUser(ctx, userId)
	if !isFound {
		return ErrUserNotFound
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
func (ulr *UserLocalMemRepository) GetCartUsers(ctx context.Context, options admin.SearchUserOptions) ([]admin.User, admin.PageMeta, error) {
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

	totalCartUsers := make([]admin.User, 0, totalElements)
	for _, value := range ulr.userCart {
		totalCartUsers = append(totalCartUsers, value)
	}

	pagedUsers := make([]admin.User, 0, options.Size)
	for i := options.Page * options.Size; i < int(math.Min(float64((options.Page+1)*options.Size), float64(totalElements))); i++ {

		pagedUsers = append(pagedUsers, totalCartUsers[i])
	}

	return pagedUsers, admin.PageMeta{
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
func getParams(options admin.SearchUserOptions, encoder *schema.Encoder) string {
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

func (ulr *UserLocalMemRepository) SetRoles(ctx context.Context, roles []admin.Role) error {
	ulr.roleCache = roles
	return nil
}
