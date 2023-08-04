package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	veracode "github.com/DanCreative/veracode-admin-plus/Veracode"
	"github.com/DanCreative/veracode-admin-plus/models"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

var Page *template.Template
var TableBody *template.Template
var Roles []models.Role
var Client *veracode.Client

// var Roles []Role = []Role{
// 	{RoleId: "cedfb4d5-c8dd-4626-bdbb-c3810f213356", RoleDescription: "Administrator", RoleName: "extadmin"},
// 	{RoleId: "25b2b166-8f46-4c66-8d52-860217f397a2", RoleDescription: "Collection Manager", RoleName: "collectionManager"},
// 	{RoleId: "488be77e-a164-43a3-b2dd-2e485b232d66", RoleDescription: "Collection Reviewer", RoleName: "collectionReviewer"},
// 	{RoleId: "7f3c7c89-c535-489f-80a7-804793e5d7e9", RoleDescription: "Creator", RoleName: "extcreator"},
// 	{RoleId: "3061ded3-55d5-4a45-a866-7ae816ea3fcc", RoleDescription: "Delete Scans", RoleName: "deletescans"},
// 	{RoleId: "c266e933-c110-416c-a486-0a9792c50545", RoleDescription: "Executive", RoleName: "extexecutive"},
// 	{RoleId: "fd49bf0b-475b-42cc-91fc-5a8adb4c6baf", RoleDescription: "Greenlight IDE User", RoleName: "greenlightideuser"},
// 	{RoleId: "9b4da96b-022a-4fa9-9ceb-3ef99f972fd7", RoleDescription: "Sandbox Administrator", RoleName: "sandboxadmin"},
// 	{RoleId: "03ccb9f7-feb1-4c46-8af3-31cae46ec153", RoleDescription: "Mitigation Approver", RoleName: "extmitigationapprover"},
// 	{RoleId: "ac32cd4e-9b36-44f9-87f4-9cadce3d7c91", RoleDescription: "Policy Administrator", RoleName: "extpolicyadmin"},
// 	{RoleId: "5c4f3f6c-3c42-4618-ade2-06127ebede95", RoleDescription: "Reviewer", RoleName: "extreviewer"},
// 	{RoleId: "df981889-59ee-449f-ac93-29da9a4df252", RoleDescription: "Sandbox User", RoleName: "sandboxuser"},
// 	{RoleId: "9fd0916c-f25c-4b7a-89e6-84d648995235", RoleDescription: "Security Insights", RoleName: "securityinsightsonly"},
// 	{RoleId: "1f59f767-33cd-4824-b616-ad10c4f985e3", RoleDescription: "Security Lead", RoleName: "extseclead"},
// 	{RoleId: "b023ec58-c6b1-43c3-ab00-88d68118d3c0", RoleDescription: "Submitter", RoleName: "extsubmitter"},
// 	{RoleId: "10ef2fe7-406e-4209-b0ce-e9deed7bf515", RoleDescription: "Team Admin", RoleName: "teamAdmin"}, // Not available for users with Admin role
// 	{RoleId: "a9cdb1a1-d6ae-4a50-b8da-c8bd12b3ffaf", RoleDescription: "Workspace Admin", RoleName: "workSpaceAdmin"},
// 	{RoleId: "189ea7d7-0628-45b8-9c68-2289f825d94b", RoleDescription: "Workspace Editor", RoleName: "workSpaceEditor"},
// // Scan Types are added by: Creator, Security Lead and Submitter
// 	{RoleId: "9824a914-3e92-4bca-806d-3d09f4c3ae75", RoleDescription: "Any Scan", RoleName: "extsubmitanyscan"},
// 	{RoleId: "c3095944-bb62-4b91-8478-1b217ecea893", RoleDescription: "Dynamic Analysis", RoleName: "extsubmitdynamicanalysis"},
// 	{RoleId: "1bb63544-1885-41d3-9b29-18a841a54275", RoleDescription: "Static Scan", RoleName: "extsubmitstaticscan"},
// 	{RoleId: "1536dee5-f7e8-454b-afae-13fbb0c1b10a", RoleDescription: "Dynamicmp Scan", RoleName: "extsubmitdynamicmpscan"},
// 	{RoleId: "ccf584b6-8b3b-4b45-bc53-12bc0fb0c9c5", RoleDescription: "Dynamic Scan", RoleName: "extsubmitdynamicscan"},
// 	{RoleId: "b6df5e84-3360-48d6-90f9-8c45784d334a", RoleDescription: "Discovery Scan", RoleName: "extsubmitdiscoveryscan"},
// }

func main() {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	logrus.SetLevel(logrus.DebugLevel)

	transport, err := veracode.NewAuthTransport(nil)
	check(err)

	Client, err = veracode.NewClient("https://api.veracode.eu/api/authn/v2", transport.Client())
	check(err)

	Roles, err = Client.GetRoles()
	check(err)

	cart := Cart{}

	router := chi.NewRouter()

	router.Get("/", IndexHandler)

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(router, "/assets", filesDir)

	router.Route("/users", func(r chi.Router) {
		r.Get("/", GetTableBody)

		r.Get("/meta", GetTableMeta)

		r.Route("/{userID}", func(r chi.Router) {
			r.Put("/", cart.PutUser)
		})
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Roles []models.Role
		Page  models.PageMeta
	}{
		Roles: Roles,
		Page:  models.PageMeta{},
	}
	err := Page.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func GetTableBody(w http.ResponseWriter, r *http.Request) {
	chTeams := make(chan any, 1)

	go Client.GetTeamsAsync(chTeams)

	r.ParseForm()
	size, err := strconv.Atoi(r.Form.Get("size"))
	if err != nil {
		size = 10
	}
	page, err := strconv.Atoi(r.Form.Get("page"))
	if err != nil {
		page = 1
	}
	users, err := Client.GetAggregatedUsers(page, size, "user")
	if err != nil {
		http.Error(w, "OOPS", 500)
	}

	for _, user := range users {
		RenderValidation(user)
	}

	var teams []models.Team

	TeamsResult := <-chTeams

	switch t := TeamsResult.(type) {
	case error:
		http.Error(w, "OOPS", 500)
	case []models.Team:
		teams = t
	}

	data := struct {
		Users []*models.User
		Teams []models.Team
	}{
		Users: users,
		Teams: teams,
	}

	TableBody.Execute(w, data)
}

func GetTableMeta(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Hit table meta")
	Page.ExecuteTemplate(w, "tableMeta", models.PageMeta{Size: 200})
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
