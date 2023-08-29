package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/DanCreative/veracode-admin-plus/utils"
	"github.com/DanCreative/veracode-admin-plus/veracode"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	table       *template.Template
	client      *veracode.Client
	cartHandler *CartHandler
}

func NewUserHandler(table *template.Template, client *veracode.Client, cartHandler *CartHandler) UserHandler {
	return UserHandler{
		table:       table,
		client:      client,
		cartHandler: cartHandler,
	}
}

func (u *UserHandler) GetTable(w http.ResponseWriter, r *http.Request) {
	chTeams := make(chan any, 1)

	go u.client.GetTeamsAsync(chTeams)

	r.ParseForm()
	size, err := strconv.Atoi(r.Form.Get("size"))
	if err != nil {
		size = 10
	}
	page, err := strconv.Atoi(r.Form.Get("page"))
	if err != nil {
		page = 1
	}
	users, meta, err := u.client.GetAggregatedUsers(page, size, "user")
	if err != nil {
		http.Error(w, "OOPS", 500)
	}

	// Cache the current query of users
	// When a user from the current query gets added to the cart,
	// The cart can store the user from the cache
	u.cartHandler.UserCache = users

	// Check if user is in the cart
	// If yes; apply cart changes to user before adding it to template data
	for _, user := range users {
		if val, ok := u.cartHandler.cart[user.UserId]; ok {
			logrus.WithFields(logrus.Fields{"Function": "GetTable"}).Tracef("Found user in cart: %s", val.EmailAddress)
			user.Roles = val.Roles
			user.Teams = val.Teams
			user.Altered = true
		}

		utils.RenderValidation(user, u.client.Roles)
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
		Roles: u.client.Roles,
		Meta:  meta,
	}

	u.table.Execute(w, data)
}
