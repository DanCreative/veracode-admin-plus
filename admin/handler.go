package admin

import (
	"net/http"
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
	users, _, err := u.userService.GetAggregatedUsers(r.Context(), SearchUserOptions{Detailed: "Yes"})
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}

	ComponentUserTable([]Team{{TeamId: "T1", TeamName: "Rivers"}, {TeamId: "T2", TeamName: "Mountains"}}, []Role{{RoleId: "R1", RoleName: "Reviewer", RoleDescription: "Reviewer"}, {RoleId: "R2", RoleName: "Administrator", RoleDescription: "Administrator"}}, users).Render(r.Context(), w)
}
