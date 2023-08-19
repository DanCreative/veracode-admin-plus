package main

import (
	"html/template"
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Cart struct {
	users   map[string]models.User
	changes *template.Template
}

func NewCart(tmpl *template.Template) Cart {
	return Cart{
		changes: tmpl,
		users:   make(map[string]models.User),
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
	w.WriteHeader(http.StatusNoContent)
}

func (c *Cart) DeleteUser(w http.ResponseWriter, r *http.Request) {

}

func (c *Cart) DeleteUsers(w http.ResponseWriter, r *http.Request) {
	c.ClearCart()
	w.WriteHeader(http.StatusNoContent)
}

func (c *Cart) GetUsers(w http.ResponseWriter, r *http.Request) {
	logrus.Debug(c.users)
	if len(c.users) > 0 {
		err := c.changes.Execute(w, c.users)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (c *Cart) SubmitCart(w http.ResponseWriter, r *http.Request) {
	errs := Client.BulkPutPartialUsers(c.users)
	if len(errs) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.ClearCart()

	w.WriteHeader(http.StatusNoContent)
}

func (c *Cart) ClearCart() {
	c.users = make(map[string]models.User)
}
