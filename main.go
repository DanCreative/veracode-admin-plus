package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

var Page *template.Template
var TableBody *template.Template

func main() {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	router := chi.NewRouter()

	router.Get("/", IndexHandler)
	router.Put("/update", UpdateHandler)

	router.Route("/users", func(r chi.Router) {
		r.Get("/", GetTableBody)
	})

	page, err := os.ReadFile("html/index.html")
	check(err)
	body, err := os.ReadFile("html/body.html")
	check(err)

	Page, err = template.New("webpage").Parse(string(page))
	check(err)
	TableBody, err = template.New("body").Parse(string(body))
	check(err)

	log.Fatal(http.ListenAndServe("localhost:8082", router))
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k, v := range r.Form {
		fmt.Printf("Permission: %s's value is: %s\n", k, v)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	roles := []Role{
		{RoleId: "cedfb4d5-c8dd-4626-bdbb-c3810f213356", RoleDescription: "Administrator"},
		{RoleId: "25b2b166-8f46-4c66-8d52-860217f397a2", RoleDescription: "Collection Manager"},
		{RoleId: "488be77e-a164-43a3-b2dd-2e485b232d66", RoleDescription: "Collection Reviewer"},
		{RoleId: "9b4da96b-022a-4fa9-9ceb-3ef99f972fd7", RoleDescription: "Sandbox Administrator"},
		{RoleId: "df981889-59ee-449f-ac93-29da9a4df252", RoleDescription: "Sandbox User"},
		{RoleId: "1f59f767-33cd-4824-b616-ad10c4f985e3", RoleDescription: "Security Lead"},
		{RoleId: "9fd0916c-f25c-4b7a-89e6-84d648995235", RoleDescription: "Security Insights"},
		{RoleId: "a9cdb1a1-d6ae-4a50-b8da-c8bd12b3ffaf", RoleDescription: "Workspace Admin"},
		{RoleId: "189ea7d7-0628-45b8-9c68-2289f825d94b", RoleDescription: "Workspace Editor"},
		{RoleId: "ac32cd4e-9b36-44f9-87f4-9cadce3d7c91", RoleDescription: "Policy Administrator"},
		{RoleId: "c266e933-c110-416c-a486-0a9792c50545", RoleDescription: "Executive"},
		{RoleId: "3061ded3-55d5-4a45-a866-7ae816ea3fcc", RoleDescription: "Delete Scans"},
		{RoleId: "fd49bf0b-475b-42cc-91fc-5a8adb4c6baf", RoleDescription: "Greenlight IDE User"},
		{RoleId: "b023ec58-c6b1-43c3-ab00-88d68118d3c0", RoleDescription: "Submitter"},
		{RoleId: "03ccb9f7-feb1-4c46-8af3-31cae46ec153", RoleDescription: "Mitigation Approver"},
		{RoleId: "5c4f3f6c-3c42-4618-ade2-06127ebede95", RoleDescription: "Reviewer"},
		{RoleId: "7f3c7c89-c535-489f-80a7-804793e5d7e9", RoleDescription: "Creator"},
		{RoleId: "10ef2fe7-406e-4209-b0ce-e9deed7bf515", RoleDescription: "Team Admin"}, // Not available for users with Admin role
		// Scan Types are added by: Creator, Security Lead and Submitter
		{RoleId: "9824a914-3e92-4bca-806d-3d09f4c3ae75", RoleDescription: "Any Scan"},
		{RoleId: "c3095944-bb62-4b91-8478-1b217ecea893", RoleDescription: "Dynamic Analysis"},
		{RoleId: "1bb63544-1885-41d3-9b29-18a841a54275", RoleDescription: "Static Scan"},
		{RoleId: "1536dee5-f7e8-454b-afae-13fbb0c1b10a", RoleDescription: "Dynamicmp Scan"},
		{RoleId: "ccf584b6-8b3b-4b45-bc53-12bc0fb0c9c5", RoleDescription: "Dynamic Scan"},
		{RoleId: "b6df5e84-3360-48d6-90f9-8c45784d334a", RoleDescription: "Discovery Scan"},
		// End of Scan Types
	}
	err := Page.Execute(w, roles)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func GetTableBody(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{
			UserName: "Dynamo@example.com",
			UserId:   "Blah",
			Roles: []Role{
				{RoleId: "cedfb4d5-c8dd-4626-bdbb-c3810f213356", IsChecked: true, IsDisabled: true},
			},
		},
	}

	time.Sleep(5 * time.Second)

	TableBody.Execute(w, users)
}
