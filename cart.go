package main

import (
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Cart struct {
	users []models.User
}

func (c *Cart) PutUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	logrus.Info(userID)
	r.ParseForm()
	logrus.Info(r.Form)
}

func (c *Cart) DeleteUser(w http.ResponseWriter, r *http.Request) {

}
