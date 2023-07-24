package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	router.Get("/", IndexHandler)
	router.Put("/update", UpdateHandler)

	log.Fatal(http.ListenAndServe("localhost:8082", router))
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k, v := range r.Form {
		fmt.Printf("Permission: %s's value is: %s\n", k, v)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	out, err := os.ReadFile("html/index.html")
	if err != nil {
		http.Error(w, "An error has occurred", http.StatusInternalServerError)
		return
	}
	w.Write(out)
}
