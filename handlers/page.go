package handlers

import (
	"html/template"
	"net/http"
)

type PageHandler struct {
	Page *template.Template
}

func (p PageHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	err := p.Page.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}
