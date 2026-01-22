package web

import (
	"net/http"
	"ocserv-users/internal"
)

func ocUsersWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "ocusers.html", nil)
}

func ocUsersPartialWebHandler(w http.ResponseWriter, r *http.Request) {
	sessions, err := internal.OcctlGetSessions()
	if err != nil {
		http.Error(w, "failed to load sessions", 500)
		return
	}

	render(w, "partials/ocusers.html", sessions)
}

func indexWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "index.html", nil)
}

func indexOcUsersPartialWebHandler(w http.ResponseWriter, r *http.Request) {
	sessions, err := internal.OcctlGetSessions()
	if err != nil {
		http.Error(w, "failed to load sessions", 500)
		return
	}

	render(w, "partials/ocusers-index.html", sessions)
}
