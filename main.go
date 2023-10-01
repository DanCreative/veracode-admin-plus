package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanCreative/veracode-admin-plus/handlers"
	"github.com/DanCreative/veracode-admin-plus/utils"
	"github.com/DanCreative/veracode-admin-plus/veracode"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

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
		"html/table/title.html",
	)
	check(err)

	// ---------------------- HANDLERS ------------------------------

	cartHandler := handlers.NewCartHandler(client)
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
		r.Delete("/filters", userHandler.DeleteFilters)
		r.Delete("/filters/{filterID}", userHandler.DeleteFilter)
	})

	router.Route("/cart", func(r chi.Router) {
		r.Post("/submit", cartHandler.SubmitCart)
		r.Delete("/", cartHandler.DeleteUsers)

		r.Route("/users", func(r chi.Router) {
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

	utils.OpenBrowser("http://localhost:8082")
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
