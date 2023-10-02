package handlers

import (
	"errors"
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/DanCreative/veracode-admin-plus/veracode"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type CartHandler struct {
	cart      map[string]models.User
	UserCache []*models.User
	client    *veracode.Client
}

// NewCartHandler creates and returns a new instance of the CartHandler model
func NewCartHandler(client *veracode.Client) CartHandler {
	return CartHandler{
		cart:   make(map[string]models.User),
		client: client,
	}
}

// PutUser handler does the following:
// * Receives the new values
// * Get user from the local cache with provided id and update said user
// * Add user to cart
func (c *CartHandler) PutUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	r.ParseForm()

	roles := []models.Role{}
	teams := []models.Team{}
	var adminTeams []string

	for k := range r.Form {
		if k == "teams" {
			for _, inputTeamId := range r.Form[k] {
				teams = append(teams, models.Team{TeamId: inputTeamId, Relationship: models.TeamRelationship{Name: "MEMBER"}})
			}

		} else if k == "admteams" {
			adminTeams = r.Form[k]

		} else {
			roles = append(roles, models.Role{RoleName: k})

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
	user.Teams = setGovernedTeams(teams, adminTeams)

	c.cart[userID] = user

	//bytes, _ := json.Marshal(user)

	logrus.WithFields(logrus.Fields{"Function": "PutUser"}).Debugf("Added user(%s) to cart: %v", userID, user)
	w.WriteHeader(http.StatusNoContent)
}

// setGovernedTeams takes a list of teams as well as a list of the ids for the
// admin teams that were selected and changes all of the teams relationships to ADMIN
func setGovernedTeams(teams []models.Team, adminTeams []string) []models.Team {
	for _, aTeamId := range adminTeams {
		for k, team := range teams {
			if team.TeamId == aTeamId {
				teams[k].Relationship.Name = "ADMIN"
			}
		}
	}
	return teams
}

// getCachedUser finds a user in the cache using a userId
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

func (c *CartHandler) ClearCart() {
	clear(c.cart)
	logrus.WithFields(logrus.Fields{"Function": "ClearCart"}).Infof("Cart cleared: %v", c.cart)
}
