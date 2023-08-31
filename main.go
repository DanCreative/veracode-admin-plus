package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanCreative/veracode-admin-plus/handlers"
	"github.com/DanCreative/veracode-admin-plus/veracode"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

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

	// ---------------------- LOGGING ------------------------------

	logrus.SetLevel(logrus.TraceLevel)

	// ---------------------- CLIENT ------------------------------

	transport, err := veracode.NewAuthTransport(nil)
	check(err)

	client, err := veracode.NewClient("https://api.veracode.eu/api/authn/v2", transport.Client())
	check(err)

	err = client.GetRoles()
	check(err)

	// ---------------------- TEMPLATES ------------------------------

	indexFile, err := os.ReadFile("html/index.html")
	check(err)

	indexTemplate, err := template.New("webpage").Parse(string(indexFile))
	check(err)

	tableTemplate, err := template.New("table").ParseFiles(
		"html/table/table.html",
		"html/table/body.html",
		"html/table/headers.html",
		"html/table/controls.html",
		"html/table/search.html",
	)
	check(err)

	cartTemplate, err := template.ParseFiles("html/cart/cart.html")
	check(err)

	// ---------------------- HANDLERS ------------------------------

	cartHandler := handlers.NewCartHandler(cartTemplate, client)
	userHandler := handlers.NewUserHandler(tableTemplate, client, &cartHandler)

	pageHandler := handlers.PageHandler{
		Page: indexTemplate,
	}

	// ---------------------- ROUTER ------------------------------

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", pageHandler.GetIndex)
	router.Route("/users", func(r chi.Router) {
		r.Get("/", userHandler.GetTable)
	})

	router.Route("/cart", func(r chi.Router) {
		r.Post("/submit", cartHandler.SubmitCart)
		r.Delete("/", cartHandler.DeleteUsers)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", cartHandler.GetUsers)
			r.Route("/{userID}", func(r chi.Router) {
				r.Put("/", cartHandler.PutUser)
			})
		})
	})

	// ---------------------- FILE SERVER ------------------------------

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(router, "/assets", filesDir)

	// ---------------------- START ------------------------------

	log.Fatal(http.ListenAndServe("localhost:8082", router))
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
