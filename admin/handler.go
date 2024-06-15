package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/DanCreative/veracode-admin-plus/common"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
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
	common.Page("/api/rest/admin/users").Render(r.Context(), w)
}

// GetUsers renders the HTML for the user table.
func (u *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	failedtmpl := func(err error, options SearchUserOptions) templ.Component {
		return ComponentUserContent(message{ShouldShow: true, IsSuccess: false, Text: err.Error()}, []Team{}, []Role{}, []User{}, options, NewPageMeta(0, 0, 0, 0, "", "", "", "", ""))
	}
	q := r.URL.Query()
	ctx := r.Context()
	options, err := getUserOptionsFromURLValues(q)
	if err != nil {
		failedtmpl(err, options).Render(ctx, w)
		return
	}

	users, pageMeta, err := u.userService.GetAggregatedUsers(r.Context(), options)
	if err != nil {
		failedtmpl(err, options).Render(ctx, w)
		return
	}

	// roles and teams are required to complete the table template.
	// "shouldRefresh" is set to true for teams, meaning that everytime the
	// table is called the teams data will be refreshed.
	roles, err := u.userService.GetRoles(r.Context(), false)
	if err != nil {
		failedtmpl(err, options).Render(ctx, w)
		return
	}

	teams, err := u.userService.GetTeams(r.Context(), true)
	if err != nil {
		failedtmpl(err, options).Render(ctx, w)
		return
	}

	ComponentUserContent(message{ShouldShow: false}, teams, roles, users, options, pageMeta).Render(ctx, w)
}

// PreSubmitUserUpdate updates the cached user's roles and teams and returns the final submission form.
func (u *UserHandler) PreSubmitUserUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userId := chi.URLParam(r, "userID")
	ctx := r.Context()

	formRoles := getRolesFromForm(r.Form)
	formTeams := getTeamsFromForm(r.Form)

	// Update User just updates locally and doesn't save to the backend
	err := u.userService.UpdateUser(ctx, userId, formRoles, formTeams)
	if err != nil {
		userErrorMessage(ctx, w, err)
		return
	}

	// roles and teams are required to complete the modal form template, however "shouldRefresh" is set to false
	// meaning that if it can be avoided, additional API calls will not be made.
	roles, err := u.userService.GetRoles(r.Context(), false)
	if err != nil {
		userErrorMessage(ctx, w, err)
		return
	}

	teams, err := u.userService.GetTeams(r.Context(), false)
	if err != nil {
		userErrorMessage(ctx, w, err)
		return
	}

	user, err := u.userService.GetUser(r.Context(), userId)
	if err != nil {
		userErrorMessage(ctx, w, err)
		return
	}

	userBytes, _ := json.Marshal(user)
	fmt.Println(string(userBytes))

	ComponentModalUserEdit(roles, teams, user).Render(r.Context(), w)
}

// Submit User updates the cached user's scan roles and admin teams and commits the changes to the backend.
func (u *UserHandler) SubmitUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userId := chi.URLParam(r, "userID")
	ctx := r.Context()

	// gets the scan type roles
	formScanRoles := getRolesFromForm(r.Form)
	formAdminTeams := getTeamsFromForm(r.Form)
	accountType := r.Form["account_type"][0]

	// roles and teams are required to complete the user's table row template, however "shouldRefresh" is set to false
	// meaning that if it can be avoided, additional API calls will not be made to get the data.
	roles, err := u.userService.GetRoles(r.Context(), false)
	if err != nil {
		ComponentMessage(message{IsSuccess: false, ShouldShow: true, Text: err.Error()}).Render(ctx, w)
		return
	}

	teams, err := u.userService.GetTeams(r.Context(), false)
	if err != nil {
		ComponentMessage(message{IsSuccess: false, ShouldShow: true, Text: err.Error()}).Render(ctx, w)
		return
	}

	returnedUser, err := u.userService.SubmitUser(ctx, userId, accountType, formScanRoles, formAdminTeams)
	if err != nil {
		ComponentMessage(message{IsSuccess: false, ShouldShow: true, Text: err.Error()}).Render(ctx, w)
		return
	}

	ComponentUserTableRow(roles, teams, returnedUser, true).Render(ctx, w)
	ComponentMessage(message{IsSuccess: true, ShouldShow: true, Text: fmt.Sprintf("User: %s was updated successfully!", userId)}).Render(ctx, w)
}

// getRolesFromForm takes a url.Values form and returns a map[string]Role
func getRolesFromForm(values url.Values) map[string]Role {
	roles := make(map[string]Role)

	if newRoleNames, ok := values["roles"]; ok {
		roles = make(map[string]Role)
		for _, RoleName := range newRoleNames {
			roles[RoleName] = Role{RoleName: RoleName}
		}
	}

	return roles
}

// getTeamsFromForm takes a url.Values form and returns a map[string]Team
func getTeamsFromForm(values url.Values) map[string]Team {
	teams := make(map[string]Team)

	if newTeamIDs, ok := values["teams"]; ok {
		teams = make(map[string]Team)
		for _, TeamID := range newTeamIDs {
			teams[TeamID] = Team{TeamId: TeamID}
		}
	}

	return teams
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

func userErrorMessage(ctx context.Context, w http.ResponseWriter, err error) {
	responseHeaders := w.Header()
	responseHeaders.Add("HX-Retarget", "#message")
	responseHeaders.Add("HX-Reswap", "outerHTML")
	ComponentMessage(message{ShouldShow: true, IsSuccess: false, Text: err.Error()}).Render(ctx, w)
}
