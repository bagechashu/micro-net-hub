package web

import (
	"embed"
	"net/http"
)

//go:embed templates/404.html
var errorTemplateFS embed.FS

// notFoundHandler handles 404 errors for undefined routes
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"RequestPath": r.URL.Path,
	}

	w.WriteHeader(http.StatusNotFound)
	renderWithLayout(w, "404.html", data)
}

// forbiddenHandler handles 403 errors
func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"RequestPath": r.URL.Path,
	}

	w.WriteHeader(http.StatusForbidden)
	renderWithLayout(w, "403.html", data)
}
