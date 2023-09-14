package handlers

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/DanCreative/veracode-admin-plus/utils"
	"github.com/DanCreative/veracode-admin-plus/veracode"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

var filterFriendlyNames = map[string]string{
	"cart":          "Cart", // Internal
	"search_term":   "Search",
	"role_id":       "Role",      // value is friendly
	"user_type":     "User Type", // value is friendly
	"login_enabled": "Login Enabled",
	"login_status":  "Login Status",
	"saml_user":     "SAML",
	"team_id":       "Team Membership", // value is friendly
}

var defaultParams = map[string]models.Filter{
	"detailed":  {CanUpdate: false, CanDelete: false, DefaultValue: "true"},
	"user_type": {CanUpdate: false, CanDelete: false, DefaultValue: "user"},
	"size":      {CanUpdate: true, CanDelete: false, DefaultValue: "10"},
	"page":      {CanUpdate: true, CanDelete: false, DefaultValue: "0"},
}

type UserHandler struct {
	table       *template.Template
	client      *veracode.Client
	cartHandler *CartHandler
	query       url.Values
}

// NewUserHandler constructs a new user handler
func NewUserHandler(table *template.Template, client *veracode.Client, cartHandler *CartHandler) UserHandler {
	q := make(url.Values)
	setDefaultParams(&q)

	return UserHandler{
		table:       table,
		client:      client,
		cartHandler: cartHandler,
		query:       q,
	}
}

// setDefaultParams sets the default values
func setDefaultParams(q *url.Values) {
	for k, v := range defaultParams {
		q.Add(k, v.DefaultValue)
	}
}

// getSearchParams takes the url values, Veracode teams and Veracode roles, and
// creates a slice of the Filter model which will be used to build the filter "pills"
// in the table template(html/table/search.html).
func getSearchParams(q url.Values, teams []models.Team, roles []models.Role) []models.Filter {
	var r []models.Filter

	for k, v := range filterFriendlyNames {
		if q.Has(k) {
			logrus.WithFields(logrus.Fields{"Function": "getSearchParams"}).Debug(k)
			filter := models.Filter{Label: k, FriendlyLabel: v, Value: q.Get(k), FriendlyValue: q.Get(k)}
			if k == "role_id" {
				for _, role := range roles {
					if role.RoleId == q.Get(k) {
						filter.FriendlyValue = role.RoleDescription
						break
					}
				}
			} else if k == "user_type" {
				switch q.Get(k) {
				case "api":
					filter.FriendlyValue = "API User"
				case "user":
					filter.FriendlyValue = "UI User"
				}
			} else if k == "team_id" {
				for _, team := range teams {
					if team.TeamId == q.Get(k) {
						filter.FriendlyValue = team.TeamName
						break
					}
				}
			}
			if p, ok := defaultParams[k]; (ok && p.CanDelete) || !ok {
				filter.CanDelete = true
			}
			r = append(r, filter)
		}
	}

	return r
}

// DeleteFilter handler deletes a single filter (if that filter can be deleted)
// This endpoint will be called when the user clicks on the x in the filter pills.
func (u *UserHandler) DeleteFilter(w http.ResponseWriter, r *http.Request) {
	filterID := chi.URLParam(r, "filterID")
	if v, ok := defaultParams[filterID]; !ok || v.CanDelete {
		u.query.Del(filterID)
	}

	u.GetTable(w, r)
}

// DeleteFilters handler deletes all of the active filters (only deletes filters that can be deleted)
// This endpoint will be called when the user clicks on the "Clear Filter" button.
func (u *UserHandler) DeleteFilters(w http.ResponseWriter, r *http.Request) {
	for k := range u.query {
		if v, ok := defaultParams[k]; !ok || v.CanDelete {
			u.query.Del(k)
		}
	}
	u.GetTable(w, r)
}

// updateQuery takes the url values from the latest call to the /users endpoint and
// updates/sets the values in the internal url.values object.
func (u *UserHandler) updateQuery(q url.Values) {
	for k, v := range q {
		if k == "page" {
			page, err := strconv.Atoi(q.Get("page"))
			if err != nil {
				page = 1
			}
			u.query.Set("page", fmt.Sprint(page-1))
			continue
		}
		if p, ok := defaultParams[k]; !ok || p.CanUpdate {
			if u.query.Has(k) {
				u.query.Set(k, v[0])
			} else {
				u.query.Add(k, v[0])
			}
		}
	}
}

// Returns all of the cart users as well as the meta data therof.
func (u *UserHandler) getCartUsers(q url.Values) ([]*models.User, models.PageMeta) {
	logrus.WithFields(logrus.Fields{"Function": "getCartUsers"}).Trace("Started")
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	totalUsersCount := len(u.cartHandler.cart)

	logrus.WithFields(logrus.Fields{"Function": "getCartUsers"}).Debugf("size: %d, page: %d, totalUsersCount: %d", size, page, totalUsersCount)
	totalUsers := make([]models.User, 0, totalUsersCount)
	for _, v := range u.cartHandler.cart {
		totalUsers = append(totalUsers, v)
	}

	var pagedUsers []*models.User
	for i := page * size; i < int(math.Min(float64((page+1)*size), float64(totalUsersCount))); i++ {

		pagedUsers = append(pagedUsers, &totalUsers[i])
	}

	meta := models.PageMeta{
		Size:          size,
		Number:        page,
		TotalElements: len(totalUsers),
		TotalPages:    int(math.Ceil(float64(totalUsersCount) / float64(size))),
	}
	logrus.WithFields(logrus.Fields{"Function": "getCartUsers"}).Debugf("pagedUsers: %v", pagedUsers)
	logrus.WithFields(logrus.Fields{"Function": "getCartUsers"}).Trace("Finished")
	return pagedUsers, meta
}

// GetTable handler makes a call to the Veracode users and teams API endpoints,
// builds and serves the table template.
func (u *UserHandler) GetTable(w http.ResponseWriter, r *http.Request) {
	chTeams := make(chan any, 1)

	go u.client.GetTeamsAsync(chTeams)

	q := r.URL.Query()
	u.updateQuery(q)
	logrus.Debug(u.query)

	var users []*models.User
	var meta models.PageMeta
	var err error

	if u.query.Has("cart") && u.query.Get("cart") == "Yes" {
		users, meta = u.getCartUsers(u.query)
	} else {
		users, meta, err = u.client.GetAggregatedUsers(u.query)
		if err != nil {
			http.Error(w, "OOPS", 500)
		}
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

	filters := getSearchParams(u.query, teams, u.client.Roles)

	meta.FirstElement = meta.Number*meta.Size + 1
	meta.LastElement = meta.Number*meta.Size + len(users)
	meta.Number += 1

	data := struct {
		Roles      []models.Role
		Users      []*models.User
		Teams      []models.Team
		Meta       models.PageMeta
		Filters    []models.Filter
		ShowCart   bool
		HasChanges bool
	}{
		Users:      users,
		Teams:      teams,
		Roles:      u.client.Roles,
		Meta:       meta,
		Filters:    filters,
		ShowCart:   u.query.Has("cart"),
		HasChanges: len(u.cartHandler.cart) > 0,
	}

	u.table.Execute(w, data)
}
