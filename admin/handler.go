package admin

import (
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
)

type UserHandler struct {
	userService UserService
}

// NewUserHandler constructs a new user handler
func NewUserHandler(service UserService) UserHandler {
	return UserHandler{
		userService: service,
	}
}

func (u *UserHandler) AdminUsersPage(w http.ResponseWriter, r *http.Request) {
	PageUsers().Render(r.Context(), w)
}

func (u *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	options, err := getUserOptionsFromURLValues(q)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}

	// fmt.Printf("url.Values: %v, options: %v", q, options)

	users, pageMeta, err := u.userService.GetAggregatedUsers(r.Context(), options)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}

	roles, err := u.userService.GetRoles(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}

	teams, err := u.userService.GetTeams(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}

	ComponentUserContent(message{ShouldShow: false}, teams, roles, users, options, pageMeta).Render(r.Context(), w)
}

// getUserOptionsFromURLValues decodes the request's url.Values into a SearchUserOptions.
// If the url.Values is empty, it sets the default values.
func getUserOptionsFromURLValues(values url.Values) (SearchUserOptions, error) {
	var decoder = schema.NewDecoder()
	var options SearchUserOptions

	if len(values) > 0 {
		err := decoder.Decode(&options, values)
		if err != nil {
			return SearchUserOptions{}, err
		}
	} else {
		// values are empty, set default values. This is most likely because its the users first time visiting this page.
		options = SearchUserOptions{
			Detailed: "Yes",
			Page:     0,
			Size:     10,
			UserType: "user",
		}
	}

	return options, nil
}
