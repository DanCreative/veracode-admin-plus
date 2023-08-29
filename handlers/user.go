package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	veracode "github.com/DanCreative/veracode-admin-plus/Veracode"
	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/DanCreative/veracode-admin-plus/utils"
)

type UserHandler struct {
	UserCache []models.User
	Cart      map[string]models.User
	Roles     []models.Role
	Table     *template.Template
	Changes   *template.Template
	Client    *veracode.Client
}

func (u *UserHandler) GetTable(w http.ResponseWriter, r *http.Request) {
	chTeams := make(chan any, 1)

	go u.Client.GetTeamsAsync(chTeams)

	r.ParseForm()
	size, err := strconv.Atoi(r.Form.Get("size"))
	if err != nil {
		size = 10
	}
	page, err := strconv.Atoi(r.Form.Get("page"))
	if err != nil {
		page = 1
	}
	users, meta, err := u.Client.GetAggregatedUsers(page, size, "user")
	if err != nil {
		http.Error(w, "OOPS", 500)
	}

	for _, user := range users {
		utils.RenderValidation(user, u.Roles)
	}

	var teams []models.Team

	TeamsResult := <-chTeams

	switch t := TeamsResult.(type) {
	case error:
		http.Error(w, "OOPS", 500)
	case []models.Team:
		teams = t
	}

	meta.FirstElement = meta.Number*meta.Size + 1
	meta.LastElement = meta.Number*meta.Size + len(users)
	meta.Number += 1

	data := struct {
		Roles []models.Role
		Users []*models.User
		Teams []models.Team
		Meta  models.PageMeta
	}{
		Users: users,
		Teams: teams,
		Roles: u.Roles,
		Meta:  meta,
	}

	u.Table.Execute(w, data)
}
