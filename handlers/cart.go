package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/DanCreative/veracode-admin-plus/veracode"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type CartHandler struct {
	cart      map[string]models.User
	changes   *template.Template
	UserCache []*models.User
	client    *veracode.Client
}

func NewCartHandler(tmpl *template.Template, client *veracode.Client) CartHandler {
	return CartHandler{
		changes: tmpl,
		cart:    make(map[string]models.User),
		client:  client,
	}
}

func (c *CartHandler) PutUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
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

	// Find user that is being put into the cart in the cached user list
	user, _ := c.getCachedUser(userID)

	var updatedRoles []models.Role

	for _, putRole := range roles {
		for _, clientRole := range c.client.Roles {
			if putRole.RoleName == clientRole.RoleName {
				updatedRoles = append(updatedRoles, clientRole)
			}
		}
	}

	user.Roles = updatedRoles
	user.Teams = teams

	c.cart[userID] = user

	bytes, _ := json.Marshal(user)

	logrus.WithFields(logrus.Fields{"Function": "PutUser"}).Debugf("Added user(%s) to cart: %v", userID, string(bytes))
	w.WriteHeader(http.StatusNoContent)
}

func (c *CartHandler) getCachedUser(userId string) (models.User, error) {
	var cacheUser models.User

	for _, v := range c.UserCache {
		if v.UserId == userId {
			cacheUser = *v
			return cacheUser, nil
		}
	}

	err := errors.New("cached user not found")
	logrus.WithFields(logrus.Fields{"Function": "getCachedUser"}).Error(err)
	return models.User{}, err
}

func (c *CartHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
}

func (c *CartHandler) DeleteUsers(w http.ResponseWriter, r *http.Request) {
	c.ClearCart()
	w.WriteHeader(http.StatusNoContent)
}

func (c *CartHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if len(c.cart) > 0 {
		err := c.changes.Execute(w, c.cart)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (c *CartHandler) SubmitCart(w http.ResponseWriter, r *http.Request) {
	errs := c.client.BulkPutPartialUsers(c.cart)
	if len(errs) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.ClearCart()

	w.WriteHeader(http.StatusNoContent)
}

func (c *CartHandler) ClearCart() {
	clear(c.cart)
}
