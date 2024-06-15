package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DanCreative/veracode-admin-plus/admin"
	"github.com/DanCreative/veracode-admin-plus/admin/backend"
	"github.com/DanCreative/veracode-admin-plus/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// Config
	homeDir, err := os.UserHomeDir()
	check(err)

	appService := config.NewApplicationConfigService(path.Join(homeDir, ".veracode"))
	appService.SetClient()

	// Repos
	veraRepo := backend.NewBasicBackendRepository(appService.GetClient)
	roleRepo := backend.NewRoleRepository(appService.GetClient)
	teamRepo := backend.NewTeamRepository(appService.GetClient)

	// Services
	userService := admin.NewUserService(veraRepo, roleRepo, teamRepo)

	// Handlers
	userHandler := admin.NewUserHandler(userService)
	settingsHandler := config.NewSettingsHandler(appService)

	// Routes : Pages
	r := chi.NewRouter()
	r.Get("/admin/users", userHandler.AdminUsersPage)
	r.Get("/settings", settingsHandler.SettingsPage)

	// Routes : API
	r.Get("/api/rest/admin/users", userHandler.GetUsers)
	r.Put("/api/rest/admin/users/{userID}", userHandler.PreSubmitUserUpdate)
	r.Put("/api/rest/admin/users/{userID}/submit", userHandler.SubmitUser)
	r.Get("/api/rest/settings", settingsHandler.GetSettings)
	r.Put("/api/rest/settings", settingsHandler.UpdateSettings)

	// File Service
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", filesDir)

	// Start up
	listener, err := net.Listen("tcp", "localhost:0")
	check(err)

	startPagePath := "/admin/users"
	_, err = appService.GetClient()
	if err != nil {
		startPagePath = "/settings"
	}

	OpenBrowser(fmt.Sprintf("http://localhost:%d%s", listener.Addr().(*net.TCPAddr).Port, startPagePath))
	log.Fatal(http.Serve(listener, r))
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

// OpenBrowser runs an OS command to open the application in the browser.
// This function currently supports: Windows, Darwin and Linux
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
