package web

import (
	"net/http"
)

// notFoundHandler handles 404 errors for undefined routes
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	renderWithLayout(w, "404.html", nil)
}

// methodNotAllowedHandler handles 405 errors for undefined routes
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusMethodNotAllowed, Response{Message: "Method Not Allowed"})
}

// forbiddenHandler handles 403 errors (kept for future use)
// func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusForbidden)
// 	renderWithLayout(w, "403.html", nil)
// }
