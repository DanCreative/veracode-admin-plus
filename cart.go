package main

import (
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Cart struct {
	users map[string]models.User
}

func NewCart() Cart {
	return Cart{
		users: make(map[string]models.User),
	}
}

func (c *Cart) PutUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	logrus.Info(userID)
	r.ParseForm()

	var roles []models.Role
	var teams []models.Team

	for k := range r.Form {
		if k != "teams" {
			roles = append(roles, models.Role{RoleName: k})
		} else {
			for _, v := range r.Form[k] {
				teams = append(teams, models.Team{TeamId: v})
			}
		}
	}

	user := models.User{
		Roles: roles,
		Teams: teams,
	}

	c.users[userID] = user
	logrus.Debugf("Added user(%s) to cart: %v", userID, user)
	err := Client.PutPartialUser(userID, user)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func (c *Cart) DeleteUser(w http.ResponseWriter, r *http.Request) {

}
